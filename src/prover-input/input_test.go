package input

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		val  any
		json string
	}{
		{
			name: "AccountState#AllZeros",
			val:  new(AccountState),
			json: `{"balance":"0x0","codeHash":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0","storageHash":"0x0000000000000000000000000000000000000000000000000000000000000000"}`,
		},
		{
			name: "AccountState#AllNonZero",
			val:  new(AccountState),
			json: `{"balance":"0x1","codeHash":"0xf81224da2c9fa1872376316fdc140b4b1e9dbf3f4579f37e9671575af143b617","code":"0x60806040523480","nonce":"0x1","storageHash":"0x6afb66cafa8dc5dd095cd08f9b8c043e2d3ff57781de43becac10c18b9ce1841","storage":{"0x5a5cf90ec2858883a6ef7a5781e4d5e5194a6324513dd38f7264f297a3dec717": "0x0000000000000000000000003cc6cdda760b79bafa08df41ecfa224f810dceb6"}}`,
		},
		{
			name: "Account#AllZeros",
			val:  new(Account),
			json: `{"balance":"0x0","codeHash":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0","storageHash":"0x0000000000000000000000000000000000000000000000000000000000000000"}`,
		},
		{
			name: "Account#AllNonZero",
			val:  new(Account),
			json: `{"balance":"0x1","codeHash":"0xf81224da2c9fa1872376316fdc140b4b1e9dbf3f4579f37e9671575af143b617","nonce":"0x1","storageHash":"0x6afb66cafa8dc5dd095cd08f9b8c043e2d3ff57781de43becac10c18b9ce1841"}`,
		},
		{
			name: "Extra#AllZeros",
			val:  new(Extra),
			json: `{}`,
		},
		{
			name: "Extra#AllNonZero",
			val:  new(Extra),
			json: `{"accessList":[{"address":"0x000000000000aaeb6d7670e522a718067333cd4e","storageKeys":["0xf81224da2c9fa1872376316fdc140b4b1e9dbf3f4579f37e9671575af143b617"]}],"committed":["0xab"],"stateDiffs":[{"address":"0x000000000000aaeb6d7670e522a718067333cd4e"}],"preState":{"0x000000000000aaeb6d7670e522a718067333cd4e":null}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var unmarshalled = reflect.New(reflect.TypeOf(test.val).Elem()).Interface()
			err := json.Unmarshal([]byte(test.json), unmarshalled)
			require.NoError(t, err)

			b, err := json.Marshal(unmarshalled)
			require.NoError(t, err)
			assert.JSONEq(t, test.json, string(b))
		})
	}
}
