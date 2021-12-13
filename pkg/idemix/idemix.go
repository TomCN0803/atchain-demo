package idemix

import (
	"fmt"

	"github.com/IBM/idemix"
	bccsp "github.com/IBM/idemix/bccsp"
	schemes "github.com/IBM/idemix/bccsp/schemes"
	"github.com/IBM/idemix/bccsp/schemes/dlog/crypto/translator/amcl"
	math "github.com/IBM/mathlib"
)

var AttributeNames = []string{"OU", "Role", "EnrollmentID", "RevocationHandle"}

// CSPWrapper wraps the idemix BCCSP implementation.
type CSPWrapper struct {
	csp schemes.BCCSP
}

func NewIdemixCSP() (*CSPWrapper, error) {
	curve := math.Curves[math.FP256BN_AMCL]
	translator := &amcl.Fp256bn{C: curve}

	csp, err := bccsp.New(NewDummyKeyStore(), curve, translator, true)
	if err != nil {
		return nil, err
	}

	return &CSPWrapper{csp: csp}, nil
}

func (c *CSPWrapper) Sign(userSK, userNymSK, userNymPK, issuerPK, credential, cri, digest []byte) ([]byte, error) {
	usk, err := c.importUserSK(userSK)
	if err != nil {
		return nil, fmt.Errorf("failed to sign the message: %w", err)
	}

	nymSK, err := c.importUserNymSK(userNymSK, userNymPK)
	if err != nil {
		return nil, fmt.Errorf("failed to sign the message: %w", err)
	}

	ipk, err := c.importIssuerPK(issuerPK)
	if err != nil {
		return nil, fmt.Errorf("failed to sign the message: %w", err)
	}

	return c.sign(usk, nymSK, ipk, credential, cri, digest)
}

func (c *CSPWrapper) NymSign(userSK, userNymSK, userNymPK, issuerPK, digest []byte) ([]byte, error) {
	usk, err := c.importUserSK(userSK)
	if err != nil {
		return nil, fmt.Errorf("failed to anonymously sign the message: %w", err)
	}

	nymSK, err := c.importUserNymSK(userNymSK, userNymPK)
	if err != nil {
		return nil, fmt.Errorf("failed to anonymously sign the message: %w", err)
	}

	ipk, err := c.importIssuerPK(issuerPK)
	if err != nil {
		return nil, fmt.Errorf("failed to anonymously sign the message: %w", err)
	}

	return c.nymSign(usk, nymSK, ipk, digest)
}

func (c *CSPWrapper) DeriveNymKeyPair(userSK, issuerPK []byte) ([]byte, []byte, error) {
	usk, err := c.importUserSK(userSK)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive anonymous secret key: %w", err)
	}

	ipk, err := c.importIssuerPK(issuerPK)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive anonymous secret key: %w", err)
	}

	nymSK, err := c.derivNymSK(usk, ipk)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive anonymous secret key: %w", err)
	}

	nymPK, err := nymSK.PublicKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive anonymous secret key: %w", err)
	}

	nymSKBytes, err := nymSK.Bytes()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive anonymous secret key: %w", err)
	}

	nymPKBytes, err := nymPK.Bytes()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive anonymous secret key: %w", err)
	}

	return nymSKBytes, nymPKBytes, nil
}

func (c *CSPWrapper) VerifySig(ou string, role int, issuerPK, revocationPK, signature, digest []byte) (bool, error) {
	ipk, err := c.importIssuerPK(issuerPK)
	if err != nil {
		return false, fmt.Errorf("failed to verify signature: %w", err)
	}

	revPK, err := c.importRevocationPK(revocationPK)
	if err != nil {
		return false, fmt.Errorf("failed to verify signature: %w", err)
	}

	return c.verifySig([]byte(ou), role, ipk, revPK, signature, digest)
}

func (c *CSPWrapper) VerifyNymSig(userNymPK, issuerPK, nymSig, digest []byte) (bool, error) {
	ipk, err := c.importIssuerPK(issuerPK)
	if err != nil {
		return false, fmt.Errorf("failed to verify anonymous signature: %w", err)
	}

	nymPK, err := c.importUserNymPK(userNymPK)
	if err != nil {
		return false, fmt.Errorf("failed to verify anonymous signature: %w", err)
	}

	return c.verifyNymSig(nymPK, ipk, nymSig, digest)
}

func (c *CSPWrapper) verifySig(
	ou []byte,
	role int,
	issuerPK, revocationPK schemes.Key,
	signature, digest []byte,
) (bool, error) {
	res, err := c.csp.Verify(
		issuerPK,
		signature,
		digest,
		&schemes.IdemixSignerOpts{
			Attributes: []schemes.IdemixAttribute{
				{Type: schemes.IdemixBytesAttribute, Value: ou},
				{Type: schemes.IdemixIntAttribute, Value: role},
				{Type: schemes.IdemixHiddenAttribute},
				{Type: schemes.IdemixHiddenAttribute},
			},
			RevocationPublicKey: revocationPK,
			EidIndex:            idemix.AttributeIndexEnrollmentId,
			RhIndex:             idemix.AttributeIndexRevocationHandle,
			Epoch:               0,
		},
	)
	if err != nil {
		return false, fmt.Errorf("failed to verify signature: %w", err)
	}

	return res, nil
}

func (c *CSPWrapper) verifyNymSig(userNymPK, issuerPK schemes.Key, nymSig, digest []byte) (bool, error) {
	res, err := c.csp.Verify(userNymPK, nymSig, digest, &schemes.IdemixNymSignerOpts{
		IssuerPK: issuerPK,
	})
	if err != nil {
		return false, fmt.Errorf("failed to verify anonymous signature: %w", err)
	}

	return res, nil
}

func (c *CSPWrapper) importRevocationPK(revocationPK []byte) (schemes.Key, error) {
	revocPK, err := c.csp.KeyImport(revocationPK, &schemes.IdemixRevocationPublicKeyImportOpts{Temporary: true})
	if err != nil {
		return nil, err
	}

	return revocPK, nil
}

func (c *CSPWrapper) importIssuerPK(issuerPKBytes []byte) (schemes.Key, error) {
	issuerPK, err := c.csp.KeyImport(issuerPKBytes, &schemes.IdemixIssuerPublicKeyImportOpts{
		Temporary:      true,
		AttributeNames: AttributeNames,
	})
	if err != nil {
		return nil, err
	}

	return issuerPK, nil
}

func (c *CSPWrapper) importUserSK(userSKBytes []byte) (schemes.Key, error) {
	userSK, err := c.csp.KeyImport(userSKBytes, &schemes.IdemixUserSecretKeyImportOpts{Temporary: true})
	if err != nil {
		return nil, err
	}

	return userSK, nil
}

func (c *CSPWrapper) importUserNymSK(userNymSK, userNymPK []byte) (schemes.Key, error) {
	nymSK, err := c.csp.KeyImport(append(userNymSK, userNymPK...), &schemes.IdemixNymKeyImportOpts{Temporary: true})
	if err != nil {
		return nil, fmt.Errorf("failed to import user anonymous secret key: %w", err)
	}

	return nymSK, nil
}

func (c *CSPWrapper) importUserNymPK(userNymPK []byte) (schemes.Key, error) {
	nymPK, err := c.csp.KeyImport(userNymPK, &schemes.IdemixNymPublicKeyImportOpts{Temporary: true})
	if err != nil {
		return nil, fmt.Errorf("failed to import user anonymous public key: %w", err)
	}

	return nymPK, nil
}

func (c *CSPWrapper) derivNymSK(userSK, issuerPK schemes.Key) (schemes.Key, error) {
	userNymSK, err := c.csp.KeyDeriv(userSK, &schemes.IdemixNymKeyDerivationOpts{
		Temporary: true,
		IssuerPK:  issuerPK,
	})
	if err != nil {
		return nil, err
	}

	return userNymSK, nil
}

func (c *CSPWrapper) sign(userSK, userNymSK, issuerPK schemes.Key, credential, cri, digest []byte) ([]byte, error) {
	attrMask := []schemes.IdemixAttribute{
		{Type: schemes.IdemixBytesAttribute},
		{Type: schemes.IdemixIntAttribute},
		{Type: schemes.IdemixHiddenAttribute},
		{Type: schemes.IdemixHiddenAttribute},
	}

	sig, err := c.csp.Sign(
		userSK,
		digest,
		&schemes.IdemixSignerOpts{
			Credential: credential,
			Nym:        userNymSK,
			IssuerPK:   issuerPK,
			Attributes: attrMask,
			EidIndex:   idemix.AttributeIndexEnrollmentId,
			RhIndex:    idemix.AttributeIndexRevocationHandle,
			Epoch:      0,
			CRI:        cri,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to sign the message: %w", err)
	}

	return sig, nil
}

func (c *CSPWrapper) nymSign(userSK, userNymSK, issuerPK schemes.Key, digest []byte) ([]byte, error) {
	sig, err := c.csp.Sign(userSK, digest, &schemes.IdemixNymSignerOpts{
		Nym:      userNymSK,
		IssuerPK: issuerPK,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to anonymously sign the message: %w", err)
	}

	return sig, nil
}
