package generate

import (
	"reflect"
	"testing"
)

// `key` is also an Asgard field name on several specs, so the keys have to be
// read from inside the ref block and nowhere else. Matching `key:` loosely is
// what made an earlier check report production Syncers for an unrelated field.
func TestRefKeysReadsOnlyTheRefBlock(t *testing.T) {
	body := `spec:
  # An Asgard field that happens to be called key, one level up.
  key: not-a-secret
  mssql:
    password:
      valueFrom:
        secretKeyRef:
          key: uof_db_password
          name: {{ include "app.appSecretName" . }}
  recipients:
    valueFrom:
      configMapKeyRef:
        key: email_recipients
        name: {{ include "app.appConfigMapName" . }}
  columns:
    - key: order_id
  apiKey:
    valueFrom:
      secretKeyRef:
        name: {{ include "app.appSecretName" . }}
        key: asgard_resource_api_key
`
	secret, config := refKeys(body)
	if want := []string{"uof_db_password", "asgard_resource_api_key"}; !reflect.DeepEqual(secret, want) {
		t.Errorf("secret keys = %q, want %q", secret, want)
	}
	if want := []string{"email_recipients"}; !reflect.DeepEqual(config, want) {
		t.Errorf("config keys = %q, want %q", config, want)
	}
}

// A templated key is a name this cannot report, and reporting `{{ ... }}` as a
// key to declare would be worse than saying nothing.
func TestRefKeysSkipsTemplatedKeys(t *testing.T) {
	body := `        secretKeyRef:
          key: {{ .Values.thing.key }}
          name: x
`
	secret, config := refKeys(body)
	if len(secret) != 0 || len(config) != 0 {
		t.Fatalf("reported a templated key: %q / %q", secret, config)
	}
}
