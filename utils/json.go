package utils

func ConvertStringArrayToByteSlices(layers []string) [][]byte {
	result := make([][]byte, 0)
	for _, layer := range layers {
		data := []byte(layer)
		result = append(result, data)
	}
	return result
}
