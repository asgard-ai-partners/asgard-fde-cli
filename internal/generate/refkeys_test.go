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

// **A key in an object the platform owns is not one to declare.**
//
// The closing message says what the engagement has to put into the release's
// own Secret. `preset-agent-hub` is not that: the platform creates it and the
// value is already in it, so declaring the key anyway writes a variable into
// the release's Secret that nothing reads - invisible to lint, render, the dry
// run and the plan, which is the third state asgard-fde-cli#114 is about.
//
// This regressed the moment the generator started reading the platform Secret:
// the CR became right and the instruction beside it started creating the
// orphan.
func TestRefKeysSkipsObjectsThePlatformOwns(t *testing.T) {
	body := `spec:
  apiKey:
    valueFrom:
      secretKeyRef:
        name: preset-agent-hub
        key: api_key
  git:
    auth:
      password:
        valueFrom:
          secretKeyRef:
            name: {{ include "app.appSecretName" . }}
            key: asgard-github-pat-password
`
	secret, _ := refKeys(body)
	if want := []string{"asgard-github-pat-password"}; !reflect.DeepEqual(secret, want) {
		t.Errorf("refKeys() = %v, want %v - api_key is the platform's and must not be declared", secret, want)
	}
}
