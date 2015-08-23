package discovery

import (
	"encoding/json"
	"testing"
)

func TestInstanceString(t *testing.T) {
	var (
		instance = Instance{
			ID:     "id",
			Name:   "name",
			Env:    "env",
			Region: "region",
		}
		json, _ = json.Marshal(instance)
	)

	if want, have := string(json), instance.String(); want != have {
		t.Fatalf("want %v, have %v", want, have)
	}
}
