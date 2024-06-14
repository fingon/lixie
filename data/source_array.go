/*
 * Author: Markus Stenberg <fingon@iki.fi>
 *
 * Copyright (c) 2024 Markus Stenberg
 *
 */

/* Static data source for testing */

package data

type ArraySource struct {
	Data   []*Log
	Chunk  int
	offset int
}

func (self *ArraySource) Load() ([]*Log, error) {
	ofs := self.offset
	got := len(self.Data)
	if self.offset >= got {
		return []*Log{}, nil
	}
	end := min(got, ofs+self.Chunk)
	self.offset = end
	return self.Data[ofs:end], nil
}
