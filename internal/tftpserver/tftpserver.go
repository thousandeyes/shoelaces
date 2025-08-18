package tftpserver

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	tftp "github.com/pin/tftp/v3"
)

type Server struct {
	core     *tftp.Server
	addr     string
	root     string
	readonly bool
	timeout  time.Duration
}

func New(addr, root string, readonly bool, timeout time.Duration) *Server {
	s := &Server{
		addr:     addr,
		root:     root,
		readonly: readonly,
		timeout:  timeout,
	}

	read := func(filename string, rf io.ReaderFrom) error {
		path := s.safeJoin(filename)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		// Advertise transfer size if known (helps some PXE ROMs).
		if fi, err := f.Stat(); err == nil {
			if ot, ok := rf.(tftp.OutgoingTransfer); ok {
				ot.SetSize(fi.Size())
			}
		}

		_, err = rf.ReadFrom(f)
		return err
	}

	var write func(string, io.WriterTo) error
	if s.readonly {
		write = nil // disable uploads
	} else {
		write = func(filename string, wt io.WriterTo) error {
			path := s.safeJoin(filename)
			// O_EXCL prevents overwriting boot loaders accidentally.
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = wt.WriteTo(f)
			return err
		}
	}

	core := tftp.NewServer(read, write)
	if s.timeout > 0 {
		core.SetTimeout(s.timeout)
	}
	s.core = core
	return s
}

func (s *Server) ListenAndServe() error {
	if s.root == "" {
		return errors.New("tftp: root directory is empty")
	}
	return s.core.ListenAndServe(s.addr) // blocks until Shutdown()
}

func (s *Server) Shutdown() { s.core.Shutdown() }

// safeJoin prevents directory traversal outside the TFTP root.
func (s *Server) safeJoin(name string) string {
	clean := filepath.Clean("/" + name)
	clean = strings.TrimPrefix(clean, "/")
	return filepath.Join(s.root, clean)
}
