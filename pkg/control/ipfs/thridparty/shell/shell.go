package shell

import (
	"context"
	"io"

	"github.com/ipfs/boxo/tar"
	shell "github.com/ipfs/go-ipfs-api"
)

func Cat(s *shell.Shell, ctx context.Context, path string, decrypt bool) (io.ReadCloser, error) {
	resp, err := s.Request("cat", path).
		Option("decrypt", decrypt).
		Send(ctx)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}

	return resp.Output, nil
}

func Get(s *shell.Shell, ctx context.Context, hash, outdir string, decrypt bool) error {

	resp, err := s.Request("get", hash).
		Option("create", true).
		Option("decrypt", decrypt).
		Send(ctx)
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.Error != nil {
		return resp.Error
	}

	extractor := &tar.Extractor{Path: outdir}
	return extractor.Extract(resp.Output)
}

func Add(s *shell.Shell, r io.Reader, onlyHash bool, cidVersion int) (string, error) {
	return s.Add(r, shell.OnlyHash(onlyHash), shell.CidVersion(cidVersion))
}
