// scheme_test.go - tests for Coconut signature scheme
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
package coconut

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jstuczyn/CoconutGo/coconut/utils"
	"github.com/jstuczyn/CoconutGo/elgamal"
	"github.com/jstuczyn/amcl/version3/go/amcl"
	Curve "github.com/jstuczyn/amcl/version3/go/amcl/BLS381"
)

// todo: simplify TestSchemeTTPKeygen

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestSchemeSetup(t *testing.T) {
	_, err := Setup(0)
	assert.Equal(t, ErrSetupParams, err, "Should not allow generating params for less than 1 attribute")

	params, err := Setup(10)
	assert.Nil(t, err)
	assert.Equal(t, 10, len(params.hs))
}

func keygenTest(t *testing.T, params *Params, sk *SecretKey, vk *VerificationKey) {
	g2 := params.g2

	assert.True(t, g2.Equals(vk.g2))
	assert.True(t, Curve.G2mul(vk.g2, sk.x).Equals(vk.alpha))
	assert.Equal(t, len(sk.y), len(vk.beta))
	for i := range vk.beta {
		assert.Equal(t, vk.beta[i], Curve.G2mul(vk.g2, sk.y[i]))
	}
}

func TestSchemeKeygen(t *testing.T) {
	params, err := Setup(10)
	assert.Nil(t, err)

	_, _, err = Keygen(&Params{G: params.G, p: params.p, g1: params.g1, g2: params.g2, hs: nil})
	assert.Equal(t, ErrKeygenParams, err, "Should not allow generating params for less than 1 attribute")

	sk, vk, err := Keygen(params)
	assert.Nil(t, err)

	keygenTest(t, params, sk, vk)
}

// nolint: gocyclo, dupl
func TestSchemeTTPKeygen(t *testing.T) {
	params, err := Setup(10)
	assert.Nil(t, err)

	_, _, err = TTPKeygen(params, 6, 5)
	assert.Equal(t, ErrTTPKeygenParams, err)

	_, _, err = TTPKeygen(params, 0, 6)
	assert.Equal(t, ErrTTPKeygenParams, err)

	_, _, err = TTPKeygen(&Params{G: params.G, p: params.p, g1: params.g1, g2: params.g2, hs: nil}, 6, 6)
	assert.Equal(t, ErrTTPKeygenParams, err)

	tests := []struct {
		t int
		n int
	}{
		{1, 6},
		{3, 6},
		{6, 6},
	}
	for _, test := range tests {
		repeat := 3
		q := 4
		params, _ := Setup(q)

		p := params.p

		sks, vks, err := TTPKeygen(params, test.t, test.n)
		assert.Nil(t, err)
		assert.Equal(t, len(sks), len(vks))

		// first check if they work as normal keys
		for i := range sks {
			keygenTest(t, params, sks[i], vks[i])
		}

		// choose random 2 subsets of t keys and ensure that when multiplied by langrage basis they converge to same value
		for i := 0; i < repeat; i++ {
			//
			// sks
			//

			indices1 := randomInts(test.t, test.n)
			sks21 := make([]*SecretKey, test.t)
			for i := range sks21 {
				sks21[i] = sks[indices1[i]-1]
			}
			// right now each point of sk has value of index + 1
			indices12 := make([]*Curve.BIG, test.t)
			l11 := make([]*Curve.BIG, test.t)
			for i, val := range indices1 {
				indices12[i] = Curve.NewBIGint(val)
			}
			for i := 0; i < test.t; i++ {
				l11[i] = utils.LagrangeBasis(i, p, indices12, 0)
			}

			// we can do it for all polynomials used for x and ys
			polys1 := make([]*Curve.BIG, q+1)
			// initialise
			for i := range polys1 {
				polys1[i] = Curve.NewBIG()
			}

			for i := range polys1 {
				for j := range sks21 {
					if i == 0 { // x
						polys1[i] = polys1[i].Plus(Curve.Modmul(l11[j], sks21[j].x, p))
					} else { // ys
						polys1[i] = polys1[i].Plus(Curve.Modmul(l11[j], sks21[j].y[i-1], p))
					}
				}
			}
			for i := range polys1 {
				polys1[i].Mod(p)
			}

			indices2 := randomInts(test.t, test.n)
			sks22 := make([]*SecretKey, test.t)
			for i := range sks22 {
				sks22[i] = sks[indices2[i]-1]
			}
			indices22 := make([]*Curve.BIG, test.t)
			l12 := make([]*Curve.BIG, test.t)
			for i, val := range indices2 {
				indices22[i] = Curve.NewBIGint(val)
			}
			for i := 0; i < test.t; i++ {
				l12[i] = utils.LagrangeBasis(i, p, indices22, 0)
			}

			polys2 := make([]*Curve.BIG, q+1)
			for i := range polys2 {
				polys2[i] = Curve.NewBIG()
			}

			for i := range polys2 {
				for j := range sks22 {
					if i == 0 { // x
						polys2[i] = polys2[i].Plus(Curve.Modmul(l12[j], sks22[j].x, p))
					} else { // ys
						polys2[i] = polys2[i].Plus(Curve.Modmul(l12[j], sks22[j].y[i-1], p))
					}
				}
			}
			for i := range polys2 {
				polys2[i].Mod(p)
				assert.Zero(t, Curve.Comp(polys1[i], polys2[i]))
			}

			// repeat the same procedure for vks (can't easily reuse code due to different types)
			//
			// vks
			//
			indices1 = randomInts(test.t, test.n)
			vks21 := make([]*VerificationKey, test.t)
			for i := range sks21 {
				vks21[i] = vks[indices1[i]-1]
			}
			// right now each point of sk has value of index + 1
			indices12 = make([]*Curve.BIG, test.t)
			l11 = make([]*Curve.BIG, test.t)
			for i, val := range indices1 {
				indices12[i] = Curve.NewBIGint(val)
			}
			for i := 0; i < test.t; i++ {
				l11[i] = utils.LagrangeBasis(i, p, indices12, 0)
			}

			// we can do it for all polynomials used for alpha and betas
			polys1v := make([]*Curve.ECP2, q+1)
			// initialise
			for i := range polys1v {
				polys1v[i] = Curve.NewECP2()
			}

			for i := range polys1v {
				for j := range vks21 {
					if i == 0 { // alpha
						polys1v[i].Add(Curve.G2mul(vks21[j].alpha, l11[j]))
					} else { // beta
						polys1v[i].Add(Curve.G2mul(vks21[j].beta[i-1], l11[j]))
					}
				}
			}

			indices2 = randomInts(test.t, test.n)
			vks22 := make([]*VerificationKey, test.t)
			for i := range sks22 {
				vks22[i] = vks[indices2[i]-1]
			}
			indices22 = make([]*Curve.BIG, test.t)
			l12 = make([]*Curve.BIG, test.t)
			for i, val := range indices2 {
				indices22[i] = Curve.NewBIGint(val)
			}
			for i := 0; i < test.t; i++ {
				l12[i] = utils.LagrangeBasis(i, p, indices22, 0)
			}

			polys2v := make([]*Curve.ECP2, q+1)
			for i := range polys2v {
				polys2v[i] = Curve.NewECP2()
			}

			for i := range polys2v {
				for j := range vks22 {
					if i == 0 { // alpha
						polys2v[i].Add(Curve.G2mul(vks22[j].alpha, l12[j]))
					} else { // beta
						polys2v[i].Add(Curve.G2mul(vks22[j].beta[i-1], l12[j]))
					}
				}
			}
			for i := range polys2v {
				assert.True(t, polys1v[i].Equals(polys2v[i]))
			}
		}
	}
}

func TestSchemeSign(t *testing.T) {
	tests := []struct {
		q     int
		attrs []string
		err   error
		msg   string
	}{
		{q: 1, attrs: []string{"Hello World!"}, err: nil,
			msg: "For single attribute sig2 should be equal to (x + m * y) * sig1"},
		{q: 3, attrs: []string{"Foo", "Bar", "Baz"}, err: nil,
			msg: "For three attributes sig2 shguld be equal to (x + m1 * y1 + m2 * y2 + m3 * y3) * sig1"},
		{q: 2, attrs: []string{"Foo", "Bar", "Baz"}, err: ErrSignParams,
			msg: "Sign should fail due to invalid param combination"},
		{q: 3, attrs: []string{"Foo", "Bar"}, err: ErrSignParams,
			msg: "Sign should fail due to invalid param combination"},
	}

	for _, test := range tests {
		params, err := Setup(test.q)
		assert.Nil(t, err)
		p := params.p

		sk, _, err := Keygen(params)
		assert.Nil(t, err)

		attrsBig := make([]*Curve.BIG, len(test.attrs))
		for i := range test.attrs {
			attrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.attrs[i])
			assert.Nil(t, err)
		}

		sig, err := Sign(params, sk, attrsBig)
		if test.err == ErrSignParams {
			assert.Equal(t, ErrSignParams, err, test.msg)
			return // everything beyond that point is UB
		}
		assert.Nil(t, err)

		t1 := Curve.NewBIGcopy(sk.x)
		for i := range sk.y {
			t1 = t1.Plus(Curve.Modmul(attrsBig[i], sk.y[i], p))
		}

		sigTest := Curve.G1mul(sig.sig1, t1)
		assert.True(t, sigTest.Equals(sig.sig2), test.msg)
	}
}

func TestSchemeVerify(t *testing.T) {
	tests := []struct {
		attrs          []string
		maliciousAttrs []string
		msg            string
	}{
		{attrs: []string{"Hello World!"}, maliciousAttrs: []string{},
			msg: "Should verify a valid signature on single public attribute"},
		{attrs: []string{"Foo", "Bar", "Baz"}, maliciousAttrs: []string{},
			msg: "Should verify a valid signature on multiple public attribute"},
		{attrs: []string{"Hello World!"}, maliciousAttrs: []string{"Malicious Hello World!"},
			msg: "Should not verify a signature when malicious attribute is introduced"},
		{attrs: []string{"Foo", "Bar", "Baz"}, maliciousAttrs: []string{"Foo2", "Bar2", "Baz2"},
			msg: "Should not verify a signature when malicious attributes are introduced"},
	}

	for _, test := range tests {
		params, err := Setup(len(test.attrs))
		assert.Nil(t, err)

		sk, vk, err := Keygen(params)
		assert.Nil(t, err)

		attrsBig := make([]*Curve.BIG, len(test.attrs))
		for i := range test.attrs {
			attrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.attrs[i])
			assert.Nil(t, err)
		}
		sig, err := Sign(params, sk, attrsBig)
		assert.Nil(t, err)
		assert.True(t, Verify(params, vk, attrsBig, sig), test.msg)

		if len(test.maliciousAttrs) > 0 {
			mAttrsBig := make([]*Curve.BIG, len(test.maliciousAttrs))
			for i := range test.maliciousAttrs {
				mAttrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.maliciousAttrs[i])
				assert.Nil(t, err)
			}
			sig2, err := Sign(params, sk, mAttrsBig)
			assert.Nil(t, err)

			assert.False(t, Verify(params, vk, attrsBig, sig2), test.msg)
			assert.False(t, Verify(params, vk, mAttrsBig, sig), test.msg)
		}
	}
}

// todo: add tests for private
func TestSchemeRandomize(t *testing.T) {
	tests := []struct {
		attrs []string
		msg   string
	}{
		{attrs: []string{"Hello World!"}, msg: "Should verify a randomized signature on single public attribute"},
		{attrs: []string{"Foo", "Bar", "Baz"}, msg: "Should verify a radomized signature on three public attribute"},
	}

	for _, test := range tests {
		params, err := Setup(len(test.attrs))
		assert.Nil(t, err)

		sk, vk, err := Keygen(params)
		assert.Nil(t, err)

		attrsBig := make([]*Curve.BIG, len(test.attrs))
		for i := range test.attrs {
			var err error
			attrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.attrs[i])
			assert.Nil(t, err)
		}
		sig, err := Sign(params, sk, attrsBig)
		assert.Nil(t, err)

		randSig := Randomize(params, sig)
		assert.True(t, Verify(params, vk, attrsBig, randSig), test.msg)
	}
}

func TestSchemeKeyAggregation(t *testing.T) {
	tests := []struct {
		attrs []string
		pp    *PolynomialPoints
		msg   string
	}{
		{attrs: []string{"Hello World!"}, pp: nil,
			msg: "Should verify a signature when single set of verification keys is aggregated (single attribute)"},
		{attrs: []string{"Foo", "Bar", "Baz"}, pp: nil,
			msg: "Should verify a signature when single set of verification keys is aggregated (three attributes)"},
		{attrs: []string{"Hello World!"}, pp: &PolynomialPoints{[]*Curve.BIG{Curve.NewBIGint(1)}},
			msg: "Should verify a signature when single set of verification keys is aggregated (single attribute)"},
		{attrs: []string{"Foo", "Bar", "Baz"}, pp: &PolynomialPoints{[]*Curve.BIG{Curve.NewBIGint(1)}},
			msg: "Should verify a signature when single set of verification keys is aggregated (three attributes)"},
	}

	for _, test := range tests {
		params, err := Setup(len(test.attrs))
		assert.Nil(t, err)

		sk, vk, err := Keygen(params)
		assert.Nil(t, err)

		attrsBig := make([]*Curve.BIG, len(test.attrs))
		for i := range test.attrs {
			attrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.attrs[i])
			assert.Nil(t, err)
		}

		sig, err := Sign(params, sk, attrsBig)
		assert.Nil(t, err)

		avk := AggregateVerificationKeys(params, []*VerificationKey{vk}, test.pp)
		assert.True(t, Verify(params, avk, attrsBig, sig), test.msg)
	}
}

// This particular test does not test the threshold property
func TestSchemeAggregateVerification(t *testing.T) {
	tests := []struct {
		attrs          []string
		authorities    int
		maliciousAuth  int
		maliciousAttrs []string
		pp             *PolynomialPoints
		t              int
		msg            string
	}{
		{attrs: []string{"Hello World!"}, authorities: 1, maliciousAuth: 0, maliciousAttrs: []string{}, pp: nil, t: 0,
			msg: "Should verify aggregated signature when only single signature was used for aggregation"},
		{attrs: []string{"Hello World!"}, authorities: 3, maliciousAuth: 0, maliciousAttrs: []string{}, pp: nil, t: 0,
			msg: "Should verify aggregated signature when three signatures were used for aggregation"},
		{attrs: []string{"Foo", "Bar", "Baz"}, authorities: 1, maliciousAuth: 0, maliciousAttrs: []string{}, pp: nil, t: 0,
			msg: "Should verify aggregated signature when only single signature was used for aggregation"},
		{attrs: []string{"Foo", "Bar", "Baz"}, authorities: 3, maliciousAuth: 0, maliciousAttrs: []string{}, pp: nil, t: 0,
			msg: "Should verify aggregated signature when three signatures were used for aggregation"},
		{attrs: []string{"Hello World!"}, authorities: 1, maliciousAuth: 2,
			maliciousAttrs: []string{"Malicious Hello World!"},
			pp:             nil,
			t:              0,
			msg:            "Should fail to verify aggregated where malicious signatures were introduced"},
		{attrs: []string{"Foo", "Bar", "Baz"}, authorities: 3, maliciousAuth: 2,
			maliciousAttrs: []string{"Foo2", "Bar2", "Baz2"},
			pp:             nil,
			t:              0,
			msg:            "Should fail to verify aggregated where malicious signatures were introduced"},

		{attrs: []string{"Hello World!"}, authorities: 1, maliciousAuth: 0,
			maliciousAttrs: []string{},
			pp:             &PolynomialPoints{[]*Curve.BIG{Curve.NewBIGint(1)}},
			t:              1,
			msg:            "Should verify aggregated signature when only single signature was used for aggregation +threshold"},
		{attrs: []string{"Hello World!"}, authorities: 3, maliciousAuth: 0,
			maliciousAttrs: []string{},
			pp:             &PolynomialPoints{[]*Curve.BIG{Curve.NewBIGint(1), Curve.NewBIGint(2), Curve.NewBIGint(3)}},
			t:              2,
			msg:            "Should verify aggregated signature when three signatures were used for aggregation +threshold"},
		{attrs: []string{"Foo", "Bar", "Baz"}, authorities: 1, maliciousAuth: 0,
			maliciousAttrs: []string{},
			pp:             &PolynomialPoints{[]*Curve.BIG{Curve.NewBIGint(1)}},
			t:              1,
			msg:            "Should verify aggregated signature when only single signature was used for aggregation +threshold"},
		{attrs: []string{"Foo", "Bar", "Baz"}, authorities: 3, maliciousAuth: 0,
			maliciousAttrs: []string{},
			pp:             &PolynomialPoints{[]*Curve.BIG{Curve.NewBIGint(1), Curve.NewBIGint(2), Curve.NewBIGint(3)}},
			t:              2,
			msg:            "Should verify aggregated signature when three signatures were used for aggregation +threshold"},
	}

	for _, test := range tests {
		params, err := Setup(len(test.attrs))
		assert.Nil(t, err)

		var sks []*SecretKey
		var vks []*VerificationKey

		if test.pp == nil {
			sks = make([]*SecretKey, test.authorities)
			vks = make([]*VerificationKey, test.authorities)
			for i := 0; i < test.authorities; i++ {
				sk, vk, err := Keygen(params)
				assert.Nil(t, err)
				sks[i] = sk
				vks[i] = vk
			}
		} else {
			sks, vks, err = TTPKeygen(params, test.t, test.authorities)
			assert.Nil(t, err)
		}

		attrsBig := make([]*Curve.BIG, len(test.attrs))
		for i := range test.attrs {
			attrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.attrs[i])
			assert.Nil(t, err)
		}

		signatures := make([]*Signature, test.authorities)
		for i := 0; i < test.authorities; i++ {
			signatures[i], err = Sign(params, sks[i], attrsBig)
			assert.Nil(t, err)
		}

		aSig := AggregateSignatures(params, signatures, test.pp)
		avk := AggregateVerificationKeys(params, vks, test.pp)

		assert.True(t, Verify(params, avk, attrsBig, aSig), test.msg)

		if test.maliciousAuth > 0 {
			msks := make([]*SecretKey, test.maliciousAuth)
			mvks := make([]*VerificationKey, test.maliciousAuth)
			for i := 0; i < test.maliciousAuth; i++ {
				sk, vk, err := Keygen(params)
				assert.Nil(t, err)
				msks[i] = sk
				mvks[i] = vk
			}

			mAttrsBig := make([]*Curve.BIG, len(test.maliciousAttrs))
			for i := range test.maliciousAttrs {
				mAttrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.maliciousAttrs[i])
				assert.Nil(t, err)
			}

			mSignatures := make([]*Signature, test.maliciousAuth)
			for i := 0; i < test.maliciousAuth; i++ {
				mSignatures[i], err = Sign(params, msks[i], mAttrsBig)
				assert.Nil(t, err)
			}

			maSig := AggregateSignatures(params, mSignatures, test.pp)
			mavk := AggregateVerificationKeys(params, mvks, test.pp)
			// todo: think of some way to test it if malicious authorities are present?
			maSig2 := AggregateSignatures(params, append(signatures, mSignatures...), test.pp)
			mavk2 := AggregateVerificationKeys(params, append(vks, mvks...), test.pp)

			assert.False(t, Verify(params, mavk, attrsBig, maSig), test.msg)
			assert.False(t, Verify(params, mavk2, attrsBig, maSig2), test.msg)

			assert.False(t, Verify(params, avk, mAttrsBig, maSig), test.msg)
			assert.False(t, Verify(params, mavk2, mAttrsBig, aSig), test.msg)

			assert.False(t, Verify(params, avk, mAttrsBig, maSig2), test.msg)
			assert.False(t, Verify(params, mavk2, mAttrsBig, maSig2), test.msg)
		}
	}
}

func TestSchemeBlindVerify(t *testing.T) {
	tests := []struct {
		q    int
		pub  []string
		priv []string
		err  error
		msg  string
	}{
		{q: 2, pub: []string{"Foo", "Bar"}, priv: []string{}, err: ErrPrepareBlindSignPrivate,
			msg: "Should not allow blindly signing messages with no private attributes"},
		{q: 1, pub: []string{}, priv: []string{"Foo", "Bar"}, err: ErrPrepareBlindSignParams,
			msg: "Should not allow blindly signing messages with invalid params"},
		{q: 2, pub: []string{}, priv: []string{"Foo", "Bar"}, err: nil,
			msg: "Should blindly sign a valid set of private attributes"},
		{q: 6, pub: []string{"Foo", "Bar", "Baz"}, priv: []string{"Foo2", "Bar2", "Baz2"}, err: nil,
			msg: "Should blindly sign a valid set of public and private attributes"},
		{q: 10, pub: []string{"Foo", "Bar", "Baz"}, priv: []string{"Foo2", "Bar2", "Baz2"}, err: nil,
			msg: "Should blindly sign a valid set of public and private attributes"}, // q > len(pub) + len(priv)
	}

	for _, test := range tests {
		params, err := Setup(test.q)
		assert.Nil(t, err)

		sk, vk, err := Keygen(params)
		assert.Nil(t, err)
		d, gamma := elgamal.Keygen(params.G)

		pubBig := make([]*Curve.BIG, len(test.pub))
		privBig := make([]*Curve.BIG, len(test.priv))

		for i := range test.pub {
			pubBig[i], err = utils.HashStringToBig(amcl.SHA256, test.pub[i])
			assert.Nil(t, err)
		}
		for i := range test.priv {
			privBig[i], err = utils.HashStringToBig(amcl.SHA256, test.priv[i])
			assert.Nil(t, err)
		}

		blindSignMats, err := PrepareBlindSign(params, gamma, pubBig, privBig)
		if len(test.priv) == 0 {
			assert.Equal(t, test.err, err)
			return
		} else if test.q < len(test.priv)+len(test.pub) {
			assert.Equal(t, test.err, err)
			return
		} else {
			assert.Nil(t, err)
		}

		// ensures len(blindSignMats.enc)+len(public_m) > len(params.hs)
		_, err = BlindSign(params, sk, blindSignMats, gamma, append(pubBig, Curve.NewBIG()))
		assert.Equal(t, ErrPrepareBlindSignParams, err, test.msg)

		incorrectGamma := Curve.NewECP()
		incorrectGamma.Copy(gamma)
		incorrectGamma.Add(Curve.NewECP()) // adds point in infinity
		// just to ensure the error is returned; proofs of knowledge are properly tested in their own test file
		_, err = BlindSign(params, sk, blindSignMats, incorrectGamma, append(pubBig, Curve.NewBIG()))
		assert.Equal(t, ErrPrepareBlindSignPrivate, err, test.msg)

		blindedSignature, err := BlindSign(params, sk, blindSignMats, gamma, pubBig)
		assert.Nil(t, err)
		sig := Unblind(params, blindedSignature, d)

		_, err = ShowBlindSignature(params, vk, sig, []*Curve.BIG{})
		assert.Equal(t, ErrShowBlindAttr, err, test.msg)

		_, err = ShowBlindSignature(params, vk, sig, append(privBig, Curve.NewBIG())) // ensures len(private_m) > len(vk.beta
		assert.Equal(t, ErrShowBlindAttr, err, test.msg)

		blindShowMats, err := ShowBlindSignature(params, vk, sig, privBig)
		assert.Nil(t, err)

		assert.True(t, BlindVerify(params, vk, sig, blindShowMats, pubBig), test.msg)
		assert.True(t, Verify(params, vk, append(privBig, pubBig...), sig), test.msg) // private attributes are revealed
	}
}

func randomInt(seen []int, max int) int {
	candidate := 1 + rand.Intn(max)
	for _, b := range seen {
		if b == candidate {
			return randomInt(seen, max)
		}
	}
	return candidate
}

// returns random (non-repetitive) q ints, > 0, < max
func randomInts(q int, max int) []int {
	ints := make([]int, q)
	seen := []int{}
	for i := range ints {
		r := randomInt(seen, max)
		ints[i] = r
		seen = append(seen, r)
	}
	return ints
}

func TestThresholdAuthorities(t *testing.T) {
	// for this purpose those randoms don't need to be securely generated
	repeat := 3
	tests := []struct {
		pub  []string
		priv []string
		t    int
		n    int
	}{
		{pub: []string{"foo", "bar"}, priv: []string{"foo2", "bar2"}, t: 1, n: 6},
		{pub: []string{"foo", "bar"}, priv: []string{"foo2", "bar2"}, t: 3, n: 6},
		{pub: []string{"foo", "bar"}, priv: []string{"foo2", "bar2"}, t: 6, n: 6},
		{pub: []string{}, priv: []string{"foo2", "bar2"}, t: 1, n: 6},
		{pub: []string{}, priv: []string{"foo2", "bar2"}, t: 3, n: 6},
		{pub: []string{}, priv: []string{"foo2", "bar2"}, t: 6, n: 6},
	}

	for _, test := range tests {

		params, err := Setup(len(test.pub) + len(test.priv))
		assert.Nil(t, err)

		d, gamma := elgamal.Keygen(params.G)

		pubBig := make([]*Curve.BIG, len(test.pub))
		privBig := make([]*Curve.BIG, len(test.priv))

		for i := range test.pub {
			pubBig[i], err = utils.HashStringToBig(amcl.SHA256, test.pub[i])
			assert.Nil(t, err)
		}
		for i := range test.priv {
			privBig[i], err = utils.HashStringToBig(amcl.SHA256, test.priv[i])
			assert.Nil(t, err)
		}

		blindSignMats, err := PrepareBlindSign(params, gamma, pubBig, privBig)
		assert.Nil(t, err)

		sks, vks, err := TTPKeygen(params, test.t, test.n)
		assert.Nil(t, err)

		// repeat the test repeat number of times to ensure it works with different subsets of keys/sigs
		for a := 0; a < repeat; a++ {
			// choose any t vks
			indices := randomInts(test.t, test.n)
			vks2 := make([]*VerificationKey, test.t)
			for i := range vks2 {
				vks2[i] = vks[indices[i]-1]
			}
			// right now each point of vk has value of index + 1
			indices12 := make([]*Curve.BIG, test.t)
			for i, val := range indices {
				indices12[i] = Curve.NewBIGint(val)
			}

			avk := AggregateVerificationKeys(params, vks2, &PolynomialPoints{indices12})

			signatures := make([]*Signature, test.n)
			for i := 0; i < test.n; i++ {
				blindedSignature, err := BlindSign(params, sks[i], blindSignMats, gamma, pubBig)
				assert.Nil(t, err)
				signatures[i] = Unblind(params, blindedSignature, d)
			}

			// and choose some other subset of t signatures
			indices2 := randomInts(test.t, test.n)
			sigs2 := make([]*Signature, test.t)
			for i := range vks2 {
				sigs2[i] = signatures[indices2[i]-1]
			}
			// right now each point of sig has value of index + 1
			indices22 := make([]*Curve.BIG, test.t)
			for i, val := range indices2 {
				indices22[i] = Curve.NewBIGint(val)
			}

			aSig := AggregateSignatures(params, sigs2, &PolynomialPoints{indices22})
			rSig := Randomize(params, aSig)

			blindShowMats, err := ShowBlindSignature(params, avk, rSig, privBig)
			assert.Nil(t, err)

			assert.True(t, BlindVerify(params, avk, rSig, blindShowMats, pubBig))
		}
	}
}

func BenchmarkSetup(b *testing.B) {
	qs := []int{1, 3, 5, 10, 20}
	for _, q := range qs {
		b.Run(fmt.Sprintf("q=%d", q), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := Setup(q)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

func BenchmarkKeygen(b *testing.B) {
	qs := []int{1, 3, 5, 10}
	for _, q := range qs {
		b.Run(fmt.Sprintf("q=%d", q), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				params, _ := Setup(q)
				b.StartTimer()
				_, _, err := Keygen(params)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

func BenchmarkTTPKeygen(b *testing.B) {
	qs := []int{1, 3, 5, 10}
	ts := []int{1, 3, 5}
	ns := []int{1, 3, 5, 10}
	for _, q := range qs {
		for _, t := range ts {
			for _, n := range ns {
				if n < t {
					continue
				}
				b.Run(fmt.Sprintf("q=%d/t=%d/n=%d", q, t, n), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						b.StopTimer()
						params, _ := Setup(q)
						b.StartTimer()
						_, _, err := TTPKeygen(params, t, n)
						if err != nil {
							panic(err)
						}
					}
				})
			}
		}
	}
}

func BenchmarkSign(b *testing.B) {
	qs := []int{1, 3, 5, 10}
	for _, q := range qs {
		b.Run(fmt.Sprintf("q=%d", q), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				params, _ := Setup(q)
				p, rng := params.p, params.G.Rng()
				pubs := make([]*Curve.BIG, q) // generate random attributes to sign
				for i := range pubs {
					pubs[i] = Curve.Randomnum(p, rng)
				}
				sk, _, _ := Keygen(params)
				b.StartTimer()
				_, err := Sign(params, sk, pubs)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

func BenchmarkPrepareBlindSign(b *testing.B) {
	privns := []int{1, 3, 5, 10}
	pubns := []int{1, 3, 5, 10}
	for _, privn := range privns {
		for _, pubn := range pubns {
			b.Run(fmt.Sprintf("pubs=%d/priv=%d", pubn, privn), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					params, _ := Setup(pubn + privn)
					p, rng := params.p, params.G.Rng()
					privs := make([]*Curve.BIG, privn) // generate random attributes to sign
					pubs := make([]*Curve.BIG, pubn)   // generate random attributes to sign

					for i := range privs {
						privs[i] = Curve.Randomnum(p, rng)
					}

					for i := range pubs {
						pubs[i] = Curve.Randomnum(p, rng)
					}

					_, gamma := elgamal.Keygen(params.G)
					b.StartTimer()
					_, err := PrepareBlindSign(params, gamma, pubs, privs)
					if err != nil {
						panic(err)
					}
				}
			})
		}
	}
}

func BenchmarkBlindSign(b *testing.B) {
	privns := []int{1, 3, 5, 10}
	pubns := []int{1, 3, 5, 10}
	for _, privn := range privns {
		for _, pubn := range pubns {
			b.Run(fmt.Sprintf("pubs=%d/priv=%d", pubn, privn), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					params, _ := Setup(pubn + privn)
					p, rng := params.p, params.G.Rng()

					privs := make([]*Curve.BIG, privn) // generate random attributes to sign
					pubs := make([]*Curve.BIG, pubn)   // generate random attributes to sign

					for i := range privs {
						privs[i] = Curve.Randomnum(p, rng)
					}

					for i := range pubs {
						pubs[i] = Curve.Randomnum(p, rng)
					}

					_, gamma := elgamal.Keygen(params.G)
					blindSignMats, _ := PrepareBlindSign(params, gamma, pubs, privs)

					sk, _, _ := Keygen(params)
					b.StartTimer()
					_, err := BlindSign(params, sk, blindSignMats, gamma, pubs)
					if err != nil {
						panic(err)
					}
				}
			})
		}
	}
}

var unblindRes *Signature

// since unblind takes constant time in relation to number of attributes,
// there is no point in embedding variable number of them into a credential
func BenchmarkUnblind(b *testing.B) {
	var sig *Signature
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		params, _ := Setup(1)
		p, rng := params.p, params.G.Rng()

		privs := []*Curve.BIG{Curve.Randomnum(p, rng)}
		pubs := []*Curve.BIG{}

		d, gamma := elgamal.Keygen(params.G)
		blindSignMats, _ := PrepareBlindSign(params, gamma, pubs, privs)

		sk, _, _ := Keygen(params)
		blindSig, _ := BlindSign(params, sk, blindSignMats, gamma, pubs)
		b.StartTimer()
		sig = Unblind(params, blindSig, d)
	}
	// it is recommended to store results in package level variables,
	// so that compiler would not try to optimize the benchmark
	unblindRes = sig
}

func BenchmarkVerify(b *testing.B) {
	qs := []int{1, 3, 5, 10}
	for _, q := range qs {
		b.Run(fmt.Sprintf("q=%d", q), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				params, _ := Setup(q)
				p, rng := params.p, params.G.Rng()
				pubs := make([]*Curve.BIG, q) // generate random attributes to sign
				for i := range pubs {
					pubs[i] = Curve.Randomnum(p, rng)
				}
				sk, vk, _ := Keygen(params)
				sig, _ := Sign(params, sk, pubs)
				b.StartTimer()
				isValid := Verify(params, vk, pubs, sig)
				if !isValid {
					panic(isValid)
				}
			}
		})
	}
}

func BenchmarkShowBlindSignature(b *testing.B) {
	privns := []int{1, 3, 5, 10}
	for _, privn := range privns {
		b.Run(fmt.Sprintf("priv=%d", privn), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				params, _ := Setup(privn)
				p, rng := params.p, params.G.Rng()

				privs := make([]*Curve.BIG, privn) // generate random attributes to sign
				pubs := []*Curve.BIG{}

				for i := range privs {
					privs[i] = Curve.Randomnum(p, rng)
				}

				d, gamma := elgamal.Keygen(params.G)
				blindSignMats, _ := PrepareBlindSign(params, gamma, pubs, privs)

				sk, vk, _ := Keygen(params)
				blindSig, _ := BlindSign(params, sk, blindSignMats, gamma, pubs)
				sig := Unblind(params, blindSig, d)
				b.StartTimer()
				_, err := ShowBlindSignature(params, vk, sig, privs)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

func BenchmarkBlindVerify(b *testing.B) {
	pubns := []int{1, 3, 5, 10}
	for _, pubn := range pubns {
		b.Run(fmt.Sprintf("pub=%d", pubn), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				params, _ := Setup(pubn + 1)
				p, rng := params.p, params.G.Rng()

				privs := []*Curve.BIG{Curve.Randomnum(p, rng)}
				pubs := make([]*Curve.BIG, pubn)

				for i := range pubs {
					pubs[i] = Curve.Randomnum(p, rng)
				}

				d, gamma := elgamal.Keygen(params.G)
				blindSignMats, _ := PrepareBlindSign(params, gamma, pubs, privs)

				sk, vk, _ := Keygen(params)
				blindSig, _ := BlindSign(params, sk, blindSignMats, gamma, pubs)
				sig := Unblind(params, blindSig, d)
				blindShowMats, _ := ShowBlindSignature(params, vk, sig, privs)

				b.StartTimer()
				isValid := BlindVerify(params, vk, sig, blindShowMats, pubs)
				if !isValid {
					panic(isValid)
				}
			}
		})
	}
}

func Example() {
	q := 5                                // number of attributes
	privM := []string{"Foo", "Bar", "42"} // private attributes
	pubM := []string{"Baz", "43"}         // public attributes

	// hash all of the attributes to BIG num:
	privMBig := make([]*Curve.BIG, len(privM))
	pubMBig := make([]*Curve.BIG, len(pubM))
	for i := range privM {
		privMBig[i], _ = utils.HashStringToBig(amcl.SHA256, privM[i])
	}
	for i := range pubM {
		pubMBig[i], _ = utils.HashStringToBig(amcl.SHA256, pubM[i])
	}

	t := 2 // threshold parameter
	n := 3 // number of authorities

	params, _ := Setup(q)
	d, gamma := elgamal.Keygen(params.G) // El-Gamal keypair

	// Generate commitment and encryption
	blindSignMats, _ := PrepareBlindSign(params, gamma, pubMBig, privMBig)

	// Generate keys for all authorities
	sks, vks, _ := TTPKeygen(params, t, n)

	// Blindly Sign attributes by each authoritiy
	blindSignatures := make([]*BlindedSignature, n)
	for i := range blindSignatures {
		blindSignatures[i], _ = BlindSign(params, sks[i], blindSignMats, gamma, pubMBig)
	}

	// Unblind all signatures
	signatures := make([]*Signature, n)
	for i := range blindSignatures {
		signatures[i] = Unblind(params, blindSignatures[i], d)
	}

	// Simple slice of indices
	pp1 := &PolynomialPoints{[]*Curve.BIG{Curve.NewBIGint(1), Curve.NewBIGint(2)}}
	pp2 := &PolynomialPoints{[]*Curve.BIG{Curve.NewBIGint(2), Curve.NewBIGint(3)}}

	// Aggregate any subset of t verification keys
	avk1 := AggregateVerificationKeys(params, vks[1:], pp2)
	avk2 := AggregateVerificationKeys(params, vks[:len(vks)-1], pp1)

	// Aggregate any subset of t credentials
	aSig1 := AggregateSignatures(params, signatures[1:], pp2)
	aSig2 := AggregateSignatures(params, signatures[:len(signatures)-1], pp1)

	// Randomize the credentials
	rSig1 := Randomize(params, aSig1)
	rSig2 := Randomize(params, aSig2)

	// Generate kappas and proofs of corectness
	blindShowMats1, _ := ShowBlindSignature(params, avk1, rSig1, privMBig)
	blindShowMats2, _ := ShowBlindSignature(params, avk2, rSig2, privMBig)
	blindShowMats3, _ := ShowBlindSignature(params, avk1, rSig2, privMBig)
	blindShowMats4, _ := ShowBlindSignature(params, avk2, rSig1, privMBig)

	// Verify credentials
	fmt.Println(BlindVerify(params, avk1, rSig1, blindShowMats1, pubMBig))
	fmt.Println(BlindVerify(params, avk2, rSig2, blindShowMats2, pubMBig))
	fmt.Println(BlindVerify(params, avk1, rSig2, blindShowMats3, pubMBig))
	fmt.Println(BlindVerify(params, avk2, rSig1, blindShowMats4, pubMBig))
}
