package bytecount

type ByteCount struct {
	Unit  string
	Count int
}

func (bc *ByteCount) getUnits() []string {
	return []string{"B", "KB", "MB", "GB", "TB"}
}

func (bc *ByteCount) Convert() {
	units := bc.getUnits()
	i := 0
	for bc.Count >= 1024 && i < len(units)-1 {
		bc.Count /= 1024
		i++
	}
	bc.Unit = units[i]
}

func (bc *ByteCount) Add(byteCount int) {
	units := bc.getUnits()
	c := byteCount
	for i := 0; i < len(units); i++ {
		if c >= 1024 {
			c /= 1024
		}
	}
	bc.Count += c
}

func (bc *ByteCount) CalcSpeed(elapsedTime int) int {
	return bc.Count / elapsedTime
}
