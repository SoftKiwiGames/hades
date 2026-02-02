package hades

import "os"

func New(stdout *os.File, stderr *os.File) *Hades {
	return &Hades{}
}

type Hades struct {

}

func (h *Hades) Run() {

}
