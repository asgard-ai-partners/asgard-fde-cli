package browser

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// The Workbench sandbox's browser (a Neko sidecar in the same pod) is driven
// over the Chrome DevTools Protocol at ASGARD_BROWSER_CDP_URL. The member sees
// that browser's visible tab when they take it over, so a page is opened by
// navigating THAT tab - the sandbox's browser skill works in it too, and a
// throwaway tab would leave the member looking at the wrong one.
//
// Navigating a tab needs one CDP command over a WebSocket. The handful of
// lines of RFC 6455 that one command needs are here rather than a dependency:
// the binary depends on almost nothing, and this is one request to a server
// in the same pod.

// EnvCDPURL is where the sandbox's browser listens for CDP.
const EnvCDPURL = "ASGARD_BROWSER_CDP_URL"

// cdpTimeout bounds the whole navigation.
const cdpTimeout = 15 * time.Second

// cdpTarget is one entry of /json/list.
type cdpTarget struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// navigateCDP opens target in the visible tab of the browser at cdpBase.
func navigateCDP(ctx context.Context, cdpBase, target string) error {
	ctx, cancel := context.WithTimeout(ctx, cdpTimeout)
	defer cancel()
	cdpBase = strings.TrimRight(cdpBase, "/")
	client := &http.Client{Timeout: cdpTimeout}

	page, err := visiblePage(ctx, client, cdpBase)
	if err != nil {
		return err
	}
	if page == nil {
		// No tab at all: make one at the page, which is then the visible one.
		req, _ := http.NewRequestWithContext(ctx, http.MethodPut, cdpBase+"/json/new?"+url.QueryEscape(target), nil)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("open a tab in the sandbox browser: %w", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("open a tab in the sandbox browser: %s", resp.Status)
		}
		return nil
	}
	ws, err := dialWebSocket(ctx, page.WebSocketDebuggerURL)
	if err != nil {
		return err
	}
	defer ws.Close()
	if err := ws.call(1, "Page.navigate", map[string]string{"url": target}); err != nil {
		return err
	}
	// Best effort: the tab is already the one the member sees in the normal
	// case, and a failure to raise it is not a failure to open the page.
	_ = ws.call(2, "Page.bringToFront", map[string]string{})
	return nil
}

// visiblePage is the tab the member sees: the first page target, as the
// browser skill's browser.contexts[0].pages[0]. Nil when there is none.
func visiblePage(ctx context.Context, client *http.Client, cdpBase string) (*cdpTarget, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cdpBase+"/json/list", nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reach the sandbox browser at %s: %w", cdpBase, err)
	}
	defer resp.Body.Close()
	var targets []cdpTarget
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&targets); err != nil {
		return nil, fmt.Errorf("read the sandbox browser's tabs: %w", err)
	}
	for i := range targets {
		t := &targets[i]
		if t.Type == "page" && t.WebSocketDebuggerURL != "" && !strings.HasPrefix(t.URL, "devtools://") {
			return t, nil
		}
	}
	return nil, nil
}

// wsConn is a client WebSocket that sends text frames and reads the
// answers, which is all one CDP call needs.
type wsConn struct {
	conn net.Conn
	r    *bufio.Reader
}

// wsGUID is RFC 6455's fixed accept-key suffix.
const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func dialWebSocket(ctx context.Context, raw string) (*wsConn, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "ws" {
		return nil, fmt.Errorf("the sandbox browser gave a CDP address this does not speak: %q", raw)
	}
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", u.Host)
	if err != nil {
		return nil, fmt.Errorf("reach the sandbox browser's tab: %w", err)
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	keyBytes := make([]byte, 16)
	_, _ = rand.Read(keyBytes)
	key := base64.StdEncoding.EncodeToString(keyBytes)
	path := u.RequestURI()
	fmt.Fprintf(conn, "GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n"+
		"Sec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n", path, u.Host, key)
	r := bufio.NewReader(conn)
	resp, err := http.ReadResponse(r, &http.Request{Method: http.MethodGet})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("upgrade the CDP connection: %w", err)
	}
	resp.Body.Close()
	sum := sha1.Sum([]byte(key + wsGUID))
	if resp.StatusCode != http.StatusSwitchingProtocols ||
		resp.Header.Get("Sec-WebSocket-Accept") != base64.StdEncoding.EncodeToString(sum[:]) {
		conn.Close()
		return nil, fmt.Errorf("the sandbox browser refused the CDP connection: %s", resp.Status)
	}
	return &wsConn{conn: conn, r: r}, nil
}

func (c *wsConn) Close() error { return c.conn.Close() }

// call sends one CDP command and waits for its answer, skipping the events
// that arrive in between.
func (c *wsConn) call(id int, method string, params any) error {
	msg, _ := json.Marshal(map[string]any{"id": id, "method": method, "params": params})
	if err := c.writeText(msg); err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	for {
		payload, err := c.readMessage()
		if err != nil {
			return fmt.Errorf("%s: %w", method, err)
		}
		var answer struct {
			ID    int `json:"id"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(payload, &answer) != nil || answer.ID != id {
			continue
		}
		if answer.Error != nil {
			return fmt.Errorf("%s: %s", method, answer.Error.Message)
		}
		return nil
	}
}

// writeText sends one masked, final text frame, as a client must.
func (c *wsConn) writeText(payload []byte) error {
	header := []byte{0x81}
	switch n := len(payload); {
	case n < 126:
		header = append(header, 0x80|byte(n))
	case n <= 0xFFFF:
		header = append(header, 0x80|126, byte(n>>8), byte(n))
	default:
		header = append(header, 0x80|127)
		header = binary.BigEndian.AppendUint64(header, uint64(n))
	}
	mask := make([]byte, 4)
	_, _ = rand.Read(mask)
	header = append(header, mask...)
	masked := make([]byte, len(payload))
	for i, b := range payload {
		masked[i] = b ^ mask[i%4]
	}
	_, err := c.conn.Write(append(header, masked...))
	return err
}

// maxCDPMessage bounds one answer; CDP answers to these two commands are tiny,
// and events that are not are skipped rather than held.
const maxCDPMessage = 16 << 20

// readMessage reads one whole message, reassembling fragments and answering
// nothing but a close.
func (c *wsConn) readMessage() ([]byte, error) {
	var msg []byte
	for {
		var h [2]byte
		if _, err := io.ReadFull(c.r, h[:]); err != nil {
			return nil, err
		}
		fin, opcode := h[0]&0x80 != 0, h[0]&0x0F
		n := uint64(h[1] & 0x7F)
		switch n {
		case 126:
			var ext [2]byte
			if _, err := io.ReadFull(c.r, ext[:]); err != nil {
				return nil, err
			}
			n = uint64(binary.BigEndian.Uint16(ext[:]))
		case 127:
			var ext [8]byte
			if _, err := io.ReadFull(c.r, ext[:]); err != nil {
				return nil, err
			}
			n = binary.BigEndian.Uint64(ext[:])
		}
		var mask []byte
		if h[1]&0x80 != 0 {
			mask = make([]byte, 4)
			if _, err := io.ReadFull(c.r, mask); err != nil {
				return nil, err
			}
		}
		if n > maxCDPMessage || uint64(len(msg))+n > maxCDPMessage {
			return nil, errors.New("the sandbox browser sent a message larger than this reads")
		}
		payload := make([]byte, n)
		if _, err := io.ReadFull(c.r, payload); err != nil {
			return nil, err
		}
		for i := range payload {
			if mask != nil {
				payload[i] ^= mask[i%4]
			}
		}
		switch opcode {
		case 0x8:
			return nil, errors.New("the sandbox browser closed the CDP connection")
		case 0x9, 0xA:
			// Ping and pong carry nothing for this.
			continue
		}
		msg = append(msg, payload...)
		if fin {
			return msg, nil
		}
	}
}
