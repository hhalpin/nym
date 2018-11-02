// elgamal.go - ElGamal encryption scheme
// Copyright (C) 2018  Jedrzej Stuczynski.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package elgamal provides primitives required by the ElGamal encryption scheme.
// It's based on Python's implementation: https://github.com/asonnino/coconut/blob/master/coconut/utils.py.
package elgamal

import (
	"github.com/jstuczyn/CoconutGo/bpgroup"

	// The named import is used to be able to easily update curve being used
	Curve "github.com/jstuczyn/amcl/version3/go/amcl/BLS381"
)

// todo: create types for public and private keys and adjust arguments accordingly (look https://godoc.org/golang.org/x/crypto/openpgp/elgamal)
// todo: possibly alternative version of Decrypt to return actual m rather than h^m
// todo: should decrypt take BpGroup argument for the sake of consistency or just remove it?

// Encryption are the two points on the G1 curve
// that represent encryption of message in form of h^m
type Encryption struct {
	c1 *Curve.ECP
	c2 *Curve.ECP
}

// C1 returns first group element of the ElGamal Encryption
func (e *Encryption) C1() *Curve.ECP {
	return e.c1
}

// C2 returns second group element of the ElGamal Encryption
func (e *Encryption) C2() *Curve.ECP {
	return e.c2
}

// NewEncryptionFromPoints wraps two points on G1 curve as ElGamal Encryption
func NewEncryptionFromPoints(c1 *Curve.ECP, c2 *Curve.ECP) *Encryption {
	return &Encryption{
		c1: c1,
		c2: c2,
	}
}

// Keygen generates private and public keys required for ElGamal encryption scheme.
// Passing coconut.Params as an argument would cause issues with cyclic dependencies,
// passing BpGroup in that case is sufficient.
func Keygen(G *bpgroup.BpGroup) (*Curve.BIG, *Curve.ECP) {
	p, g1, rng := G.Order(), G.Gen1(), G.Rng()

	d := Curve.Randomnum(p, rng)
	gamma := Curve.G1mul(g1, d)
	return d, gamma
}

// Encrypt encrypts the given message in the form of h^m,
// where h is a point on the G1 curve using the given public key.
// The random k is returned alongside the encryption
// as it is required by the Coconut Scheme to create proofs of knowledge.
func Encrypt(G *bpgroup.BpGroup, gamma *Curve.ECP, m *Curve.BIG, h *Curve.ECP) (*Encryption, *Curve.BIG) {
	p, g1, rng := G.Order(), G.Gen1(), G.Rng()

	k := Curve.Randomnum(p, rng)
	a := Curve.G1mul(g1, k)
	b := Curve.G1mul(gamma, k) // b = (k * gamma)
	b.Add(Curve.G1mul(h, m))   // b = (k * gamma) + (m * h)

	return &Encryption{a, b}, k
}

// Decrypt takes the ElGamal encryption of a message and returns a point on the G1 curve
// that represents original h^m.
func Decrypt(G *bpgroup.BpGroup, d *Curve.BIG, enc *Encryption) *Curve.ECP {
	dec := Curve.NewECP()
	dec.Copy(enc.c2)
	dec.Sub(Curve.G1mul(enc.c1, d))
	return dec
}
