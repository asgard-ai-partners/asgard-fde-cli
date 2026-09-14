package generate

import (
	"fmt"
	"sort"
	"strings"
)

// The DataConnector classes, read off asgard-kube `crd/asgard-ai.com_dataconnectors.yaml`
// at `cbd8d70` on 2026-09-07: the fields each class declares, which of them the
// CRD requires, and which are credentials.
//
// **The generator wrote two of the nine.** `--db-class` accepted postgres and
// mssql, and every other class shares nothing but `host` with them - salesforce
// has no port and no user, athena has neither host nor password, netsuite
// authenticates with a certificate. An FDE whose customer runs Oracle got a
// postgres skeleton and had to work out the difference from the CRD.
//
// Port defaults are the ones `.agents/skills/db-query/scripts/connectors.py`
// already recorded from real connections, not conventional guesses.
type dbClass struct {
	// values are the fields whose value is a chart value: the coordinates.
	values []dbField
	// secrets are the fields that may only ever be a secretKeyRef. The key is
	// suffixed onto the connector's values key.
	secrets []dbField
	// note is what the class needs that the shape cannot say.
	note string
}

type dbField struct {
	name    string
	def     string // default for the values block; empty means ""
	numeric bool
	comment string
}

var dbClasses = map[string]dbClass{
	"postgres": {
		values: []dbField{
			{name: "host"}, {name: "port", def: "5432", numeric: true},
			{name: "user"}, {name: "database"},
			{name: "sslMode", def: "disable", comment: "optional; the only class with it"},
		},
		secrets: []dbField{{name: "password"}},
	},
	"mysql": {
		values: []dbField{
			{name: "host"}, {name: "port", def: "3306", numeric: true},
			{name: "user"}, {name: "database"},
		},
		secrets: []dbField{{name: "password"}},
	},
	"mssql": {
		values: []dbField{
			{name: "host"}, {name: "port", def: "1433", numeric: true},
			{name: "user"}, {name: "database"},
			{name: "instance", comment: "optional; a named SQL Server instance"},
		},
		secrets: []dbField{{name: "password"}},
	},
	"oracle": {
		values: []dbField{
			{name: "host"}, {name: "port", def: "1521", numeric: true},
			{name: "user"},
			{name: "serviceName", comment: "exactly one of serviceName or sid - the CRD refuses both and refuses neither"},
		},
		secrets: []dbField{{name: "password"}},
		note:    "Oracle takes serviceName **or** sid, exactly one. This writes serviceName; swap the key if the customer gave you an SID.",
	},
	"hana": {
		values: []dbField{
			{name: "host"}, {name: "port", def: "30015", numeric: true},
			{name: "user"},
		},
		secrets: []dbField{{name: "password"}},
		note:    "**`db-query` cannot reach HANA**: SAP does not distribute `hdbcli` openly, so introspect it with whatever the customer's own DBAs use. The platform reads it fine - see the db-query skill's `references/connectors.md`.",
	},
	"salesforce": {
		values:  []dbField{{name: "host"}},
		secrets: []dbField{{name: "consumerKey"}, {name: "consumerSecret"}},
		note:    "No port and no user: a connected app's consumer key and secret are the whole credential.",
	},
	"netsuite": {
		values: []dbField{
			{name: "host"},
			{name: "certificateId", comment: "the certificate's id in NetSuite, not the key itself"},
			{name: "signatureAlgorithm", def: "PS256"},
		},
		secrets: []dbField{{name: "consumerKey"}, {name: "privateKeyPem"}},
		note:    "Certificate authentication: the PEM is a secret, the certificate **id** is a coordinate. `signatureAlgorithm` has no default in the CRD.",
	},
	"trino": {
		values: []dbField{
			{name: "host"}, {name: "port", def: "443", numeric: true},
			{name: "scheme", def: "https"}, {name: "user"},
		},
		secrets: []dbField{
			{name: "password", comment: "optional"},
			{name: "jwtAccessToken", comment: "optional; use this or password, not both"},
			{name: "sslCertificatePem", comment: "optional"},
		},
		note: "`scheme` is required and has no default in the CRD. Delete whichever credentials the cluster does not use - all three are optional and an empty secretKeyRef is worse than an absent one.",
	},
	"athena": {
		values: []dbField{
			{name: "region"},
			{name: "outputLocation", comment: "must start with s3:// - the CRD enforces it"},
			{name: "workGroup", comment: "optional"},
		},
		secrets: []dbField{{name: "accessKeyId"}, {name: "secretAccessKey"}},
		note:    "No host, no port, no database: a region, an S3 output location and an IAM key pair.",
	},
}

// DBClasses names every class --db-class accepts, sorted.
func DBClasses() []string {
	out := make([]string, 0, len(dbClasses))
	for k := range dbClasses {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ValidDBClass reports whether the CRD has this class.
func ValidDBClass(name string) bool {
	_, ok := dbClasses[name]
	return ok
}

// dbSpec renders the class block of a DataConnector's spec.
func dbSpec(class, valuesKey, chart string) string {
	c := dbClasses[class]
	var b strings.Builder
	fmt.Fprintf(&b, "  %s:\n", class)
	for _, f := range c.values {
		if f.comment != "" {
			fmt.Fprintf(&b, "    # %s\n", f.comment)
		}
		if f.numeric {
			fmt.Fprintf(&b, "    %s: {{ .Values.%sDB.%s }}\n", f.name, valuesKey, f.name)
		} else {
			fmt.Fprintf(&b, "    %s: {{ .Values.%sDB.%s | quote }}\n", f.name, valuesKey, f.name)
		}
	}
	for _, f := range c.secrets {
		if f.comment != "" {
			fmt.Fprintf(&b, "    # %s\n", f.comment)
		}
		fmt.Fprintf(&b, "    %s:\n      valueFrom:\n        secretKeyRef:\n          key: %s_db_%s\n          name: {{ include \"%s.appSecretName\" . }}\n",
			f.name, valuesKey, strings.ToLower(f.name), chart)
	}
	return b.String()
}

// dbValues renders the values.yaml keys the class block reads.
func dbValues(class, display, valuesKey string) string {
	c := dbClasses[class]
	var b strings.Builder
	fmt.Fprintf(&b, "\n# %s (%s)\n%sDB:\n", display, class, valuesKey)
	for _, f := range c.values {
		switch {
		case f.numeric:
			fmt.Fprintf(&b, "  %s: %s\n", f.name, f.def)
		case f.def != "":
			fmt.Fprintf(&b, "  %s: %q\n", f.name, f.def)
		default:
			fmt.Fprintf(&b, "  %s: \"\"\n", f.name)
		}
	}
	return b.String()
}

// dbNote is what the class needs that its shape cannot say, or "".
func dbNote(class string) string { return dbClasses[class].note }
