package consul

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/micro/go-config/encoder"
)

func makeMap(e encoder.Encoder, kv api.KVPairs, stripPrefix string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	for _, v := range kv {
		// remove prefix if non empty, and ensure leading / is removed as well
		vkey := strings.TrimPrefix(strings.TrimPrefix(v.Key, stripPrefix), "/")
		// split on prefix
		keys := strings.Split(vkey, "/")

		var vals interface{}
		if len(v.Value) > 0 {
			if err := e.Decode(v.Value, &vals); err != nil {
				return nil, fmt.Errorf("faild decode value. path: %s, error: %s", vkey, err)
			}
		}

		// set data for first iteration
		kvals := data

		// iterate the keys and make maps
		for i, k := range keys {
			kval, ok := kvals[k].(map[string]interface{})
			if !ok {
				// create next map
				kval = make(map[string]interface{})
				// set it
				kvals[k] = kval
			}

			// last key: write vals
			if l := len(keys) - 1; i == l {
				kvals[k] = vals
				break
			}

			// set kvals for next iterator
			kvals = kval
		}

	}

	return data, nil
}
