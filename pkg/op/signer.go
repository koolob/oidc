package op

import (
	"context"
	"errors"

	"github.com/caos/logging"
	"gopkg.in/square/go-jose.v2"
)

type Signer interface {
	Health(ctx context.Context) error
	Signer() jose.Signer
	SignatureAlgorithm() jose.SignatureAlgorithm
}

type tokenSigner struct {
	signer  jose.Signer
	storage AuthStorage
	alg     jose.SignatureAlgorithm
}

func NewSigner(ctx context.Context, storage AuthStorage, keyCh <-chan jose.SigningKey, signOpt *jose.SignerOptions) Signer {
	s := &tokenSigner{
		storage: storage,
	}

	go s.refreshSigningKey(ctx, keyCh, signOpt)

	return s
}

func (s *tokenSigner) Health(_ context.Context) error {
	if s.signer == nil {
		return errors.New("no signer")
	}
	if string(s.alg) == "" {
		return errors.New("no signing algorithm")
	}
	return nil
}

func (s *tokenSigner) Signer() jose.Signer {
	return s.signer
}

func (s *tokenSigner) refreshSigningKey(ctx context.Context, keyCh <-chan jose.SigningKey, signOpt *jose.SignerOptions) {
	for {
		select {
		case <-ctx.Done():
			return
		case key := <-keyCh:
			s.alg = key.Algorithm
			if key.Algorithm == "" || key.Key == nil {
				s.signer = nil
				logging.Log("OP-DAvt4").Warn("signer has no key")
				continue
			}
			var err error
			s.signer, err = jose.NewSigner(key, signOpt)
			if err != nil {
				logging.Log("OP-pf32aw").WithError(err).Error("error creating signer")
				continue
			}
			logging.Log("OP-agRf2").Info("signer exchanged signing key")
		}
	}
}

func (s *tokenSigner) SignatureAlgorithm() jose.SignatureAlgorithm {
	return s.alg
}
