package charger

import (
	"testing"

	"github.com/andig/evcc/api"
)

func TestKeba(t *testing.T) {
	var wb api.Charger
	wb, err := NewKeba("foo", RFID{}, 0)
	if err != nil {
		t.Error(err)
	}

	if _, ok := wb.(api.MeterEnergy); !ok {
		t.Error("missing MeterEnergy interface")
	}

	if _, ok := wb.(api.MeterCurrent); !ok {
		t.Error("missing MeterCurrent interface")
	}

	if _, ok := wb.(api.ChargeRater); !ok {
		t.Error("missing ChargeRater interface")
	}
}
