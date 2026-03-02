package divergence

import (
	"bytes"
	"encoding/json"

	"github.com/wI2L/jsondiff"
)

func DiffBodies(live, shadow []byte) ([]BodyDiffEntry, error) {
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
	return []BodyDiffEntry{}, err
}
response := []BodyDiffEntry{}
for _,v := range patch {
	response = append(response, BodyDiffEntry{
		Op: v.Type,
		Path: v.Path,
		LiveValue: v.Value,
		ShadowValue: v.OldValue,
	})

}
return response, nil
}