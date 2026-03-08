package divergence

import (
	"bytes"
	"encoding/json"

	"github.com/Dubjay/specter/internal/types"
	"github.com/wI2L/jsondiff"
)

func DiffBodies(live, shadow []byte) ([]types.BodyDiffEntry, error) {
	patch, err := jsondiff.CompareJSON(
    shadow,
    live,
    jsondiff.UnmarshalFunc(func(b []byte, v any) error {
        dec := json.NewDecoder(bytes.NewReader(b))
        dec.UseNumber()
        return dec.Decode(v)
    }),
)
if err != nil {
	return []types.BodyDiffEntry{}, err
}
response := []types.BodyDiffEntry{}
for _,v := range patch {
	response = append(response, types.BodyDiffEntry{
		Op: v.Type,
		Path: v.Path,
		LiveValue: v.Value,
		ShadowValue: v.OldValue,
	})

}
return response, nil
}