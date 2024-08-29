package bytecount

type ByteCount struct {
	Unit   string
	UnitID int
	Count  int
}

func (bc *ByteCount) getUnits() []string {
	return []string{"B", "KB", "MB", "GB", "TB"}
}

func (bc *ByteCount) getPrevUnit() string {
	return bc.getUnits()[bc.UnitID-1]
}

func (bc *ByteCount) getNextUnit() string {
	return bc.getUnits()[bc.UnitID+1]
}

// byteCount will always be in bytes. Used as new written bytes into the file
func (bc *ByteCount) Convert(byteCount int) {
	units := bc.getUnits()
	c := byteCount
	for i := 0; i < bc.UnitID; i++ {
		c /= 1024
	}
	bc.Count += c
	for bc.Count >= 1024 {
		bc.Count /= 1024
		bc.UnitID++
		bc.Unit = units[bc.UnitID]
	}
}

func (bc *ByteCount) CalcSpeed(elapsedTime int) int {
	return bc.Count / elapsedTime
}

// func ConvertBytes(b int64) string {
// 	units := []string{"B", "KB", "MB", "GB", "TB"}
// 	values := make([]int64, len(units))
// 	values[0] = b
// 	for i := 1; i < len(units); i++ {
// 		values[i] = values[i-1] / 1024
// 	}
// 	for i := len(units) - 1; i >= 0; i-- {
// 		if values[i] > 0 {
// 			return fmt.Sprintf("%d %s", values[i], units[i])
// 		}
// 	}
// 	return ""
// }
