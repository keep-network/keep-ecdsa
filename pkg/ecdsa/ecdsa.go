// Package ecdsa implements ECDSA signing it uses standards described in [SEC 1].
//
//   [SEC 1]: Standards for Efficient Cryptography, SEC 1: Elliptic Curve
//     Cryptography, Certicom Research, https://www.secg.org/sec1-v2.pdf
package ecdsa

import (
	cecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// Signature holds a signature in a form of two big.Int `r` and `s` values and a
// recovery ID value in {0, 1, 2, 3}.
//
// The signature is chain-agnostic. Some chains (e.g. Ethereum and BTC) requires
// `v` to start from 27. Please consult the documentation about what the
// particular chain expects.
type Signature struct {
	R          *big.Int
	S          *big.Int
	RecoveryID int
}

// CalculateSignature returns an signature over provided hash, calculated
// with Signer's private key. Signature is returned in `(r, s, v)` form.
func (s *Signer) CalculateSignature(rand io.Reader, hash []byte) (*Signature, error) {
	sigR, sigS, err := cecdsa.Sign(rand, s.privateKey, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate ECDSA signature: [%v]", err)
	}

	// TODO: In the future we could recover `recoverID` value only if it is required by
	// the signature requestor, if not we will return just (r,s) values.
	recoveryID, err := s.findRecoveryID(sigR, sigS, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to find recovery ID: [%v]", err)
	}

	return &Signature{
		R:          sigR,
		S:          sigS,
		RecoveryID: recoveryID,
	}, nil
}

// findRecoveryID finds recovery ID for the signature. Recovery ID is a value used
// in bitcoin and ethereum signatures to determine public key which is related
// to the signer.
//
// Signature in a form `(r, s)` contains `r` value which is a `x` cooridante of
// the point `R` representing the public key of the signer. We use public key
// recovery to get a missing `y` coordinate for the `R` point. The curve we use
// for signing has up to 4 possible points for the given `x` coordinate. We use
// recovery ID to point which of them is the one we used for signing.
//
// Internal function recoverPublicKeyFromSignature recovers a public key from the
// signature's R and S values for the given message hash. Based on the algorithm
// described in section 4.1.6 of [SEC 1]. It handles the inner loop of the
// algorithm from point 1.6 based on the `iteration` parameter. It decides to
// select `R` or `-R` value based on oddness of the y cooridnate. This is
// consistent with solution implemented by btcd.
func (s *Signer) findRecoveryID(sigR, sigS *big.Int, hash []byte) (int, error) {
	recoverPublicKeyFromSignature := func(iteration int) (*PublicKey, error) {
		curve := s.Curve()
		j := iteration / 2

		// 1.1 Calculate x coordinate of the R point.
		// x = r + (j * n)
		Rx := new(big.Int).Add(
			sigR,
			new(big.Int).Mul(
				big.NewInt(int64(j)),
				curve.Params().N,
			),
		)

		if Rx.Cmp(curve.Params().P) != -1 {
			return nil, fmt.Errorf("calculated Rx is larger than curve P")
		}

		// 1.3 Estimate y coordinate of the R point. For each x cooridnate there are
		// two possible points on the elliptic curve - `R` and `-R`.

		// `calculateY` always returns `y` coordinate of `R` point. We calculate `-R`
		// point later if it is required.
		Ry := calculateY(curve, Rx)
		if Ry == nil {
			return nil, fmt.Errorf("failed to calculate y")
		}

		// We check a `negativeR` flag to check which point we should use in our
		// calculation.
		// We compare it with oddness of the y coordinate to match btcec solution.
		oddIteration := iteration%2 == 1
		if oddIteration != isOdd(Ry) {
			Ry = new(big.Int).Mod(
				new(big.Int).Neg(Ry),
				curve.Params().P,
			)
		}

		// Validate found point.
		if !curve.IsOnCurve(Rx, Ry) {
			return nil, fmt.Errorf("point is not on curve")
		}

		// 1.5 Calculate `e` from message using the same algorithm as ecdsa signature
		// calculation.
		e := hashToInt(curve, hash)

		// 1.6.1 Calculate a candidate public key.
		// Q = (r^-1) * ( (s * R) - (e * G))
		rInverse := new(big.Int).ModInverse(sigR, curve.Params().N) // (r^-1)

		sRx, sRy := curve.ScalarMult(Rx, Ry, sigS.Bytes()) // (s * R)

		// - (e * G)
		minusE := new(big.Int).Mod(
			new(big.Int).Neg(e),
			curve.Params().N)
		minuseGx, minuseGy := curve.ScalarBaseMult(minusE.Bytes())

		// (s * R) - (e * G)
		addedX, addedY := curve.Add(
			sRx, sRy,
			minuseGx, minuseGy,
		)

		Qx, Qy := curve.ScalarMult( // (r^-1) * ( (s * R) - (e * G))
			addedX, addedY,
			rInverse.Bytes(),
		)

		// We don't perform validation mentioned in point 1.6.2, because we expect
		// it to be performed inside of the loop calling this function.

		return &PublicKey{Curve: curve, X: Qx, Y: Qy}, nil
	}

	// `h` is a co-factor of the elliptic curve, for the curve we are using (scp256k1)
	// the co-factor is equal `1`
	h := 1

	// We iterate over `2*(h+1) = 4` possible recovery ID values. For given
	// signature there are 4 public keys against which the signature is valid.
	// Here we check which recovery ID will correspond with signer's public key.
	for i := 0; i < 2*(h+1); i++ {
		publicKey, err := recoverPublicKeyFromSignature(i)
		if err != nil {
			// If key recovery is executed in appropriate order we we don't expect
			// the error to happen.
			// TODO: Change to logger.Warnf after logging is merged to master.
			fmt.Printf(
				"failed to recover public key in iteration [%d]: [%v]",
				i,
				err,
			)
			continue
		}

		// Check if recovered public key matches signer's public key. If not
		// continue iteration for a next possible recovery ID.
		if publicKey.X.Cmp(s.PublicKey().X) == 0 &&
			publicKey.Y.Cmp(s.PublicKey().Y) == 0 {
			return i, nil
		}
	}

	return -1, fmt.Errorf("failed to find recovery ID")
}

// calculateY calculates `y` coordinate for `x` curve point coordinate using the provided
// elliptic curve implementation. It expects the elliptic curve to be a
// short-form Weierstrass curve with `a = 0`, defined by the equation `y² = x³ + b`.
// `b` is a constant of the curve equation, specific for the particular curve.
func calculateY(curve *secp256k1.BitCurve, x *big.Int) *big.Int {
	// x³
	x3 := new(big.Int).Exp(x, big.NewInt(3), big.NewInt(0))

	// x³ + b
	y2 := new(big.Int).Add(x3, curve.Params().B)

	// solve y² = x³ + b
	return x3.ModSqrt(y2, curve.Params().P)
}

// This code is borrowed from Golang's crypto/ecdsa/ecdsa.go.
func hashToInt(c elliptic.Curve, hash []byte) *big.Int {
	orderBits := c.Params().N.BitLen()
	orderBytes := (orderBits + 7) / 8
	if len(hash) > orderBytes {
		hash = hash[:orderBytes]
	}

	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return ret
}

func isOdd(a *big.Int) bool {
	return a.Bit(0) == 1
}
