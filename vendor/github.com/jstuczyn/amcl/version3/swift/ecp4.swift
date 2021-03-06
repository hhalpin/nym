/*
	Licensed to the Apache Software Foundation (ASF) under one
	or more contributor license agreements.  See the NOTICE file
	distributed with this work for additional information
	regarding copyright ownership.  The ASF licenses this file
	to you under the Apache License, Version 2.0 (the
	"License"); you may not use this file except in compliance
	with the License.  You may obtain a copy of the License at
	
	http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing,
	software distributed under the License is distributed on an
	"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
	KIND, either express or implied.  See the License for the
	specific language governing permissions and limitations
	under the License.
*/
//
//  ecp4.swift
//
//  Created by Michael Scott on 07/07/2015.
//  Copyright (c) 2015 Michael Scott. All rights reserved.
//

/* AMCL Weierstrass elliptic curve functions over FP4 */

final public class ECP4 {
    private var x:FP4
    private var y:FP4
    private var z:FP4
 //   private var INF:Bool
    
    /* Constructor - set self=O */
    init()
    {
    //    INF=true
        x=FP4(0)
        y=FP4(1)
        z=FP4(0)
    }
    /* Test self=O? */
    public func is_infinity() -> Bool
    {
    //    if INF {return true}
        return x.iszilch() && z.iszilch()
    }
    /* copy self=P */
    public func copy(_ P:ECP4)
    {
        x.copy(P.x)
        y.copy(P.y)
        z.copy(P.z)
    //    INF=P.INF
    }
    /* set self=O */
    func inf() {
    //    INF=true
        x.zero()
        y.one()
        z.zero()
    }

    /* Conditional move of Q to P dependant on d */
    func cmove(_ Q:ECP4,_ d:Int)
    {
        x.cmove(Q.x,d);
        y.cmove(Q.y,d);
        z.cmove(Q.z,d);
    /*
        var bd:Bool
        if d==0 {bd=false}
        else {bd=true}
        INF = (INF != ((INF != Q.INF) && bd)) */
    }
    
    /* return 1 if b==c, no branching */
    private static func teq(_ b:Int32,_ c:Int32) -> Int
    {
        var x=b^c
        x-=1  // if x=0, x now -1
        return Int((x>>31)&1)
    }
    /* Constant time select from pre-computed table */
    func select(_ W:[ECP4],_ b:Int32)
    {
        let MP=ECP4()
        let m=b>>31
        var babs=(b^m)-m
        
        babs=(babs-1)/2
    
        cmove(W[0],ECP4.teq(babs,0)) // conditional move
        cmove(W[1],ECP4.teq(babs,1))
        cmove(W[2],ECP4.teq(babs,2))
        cmove(W[3],ECP4.teq(babs,3))
        cmove(W[4],ECP4.teq(babs,4))
        cmove(W[5],ECP4.teq(babs,5))
        cmove(W[6],ECP4.teq(babs,6))
        cmove(W[7],ECP4.teq(babs,7))
    
        MP.copy(self)
        MP.neg()
        cmove(MP,Int(m&1))
    }

    /* Test if P == Q */
    func equals(_ Q:ECP4) -> Bool
    {
    //    if is_infinity() && Q.is_infinity() {return true}
    //    if is_infinity() || Q.is_infinity() {return false}
    
        let a=FP4(x)                            // *****
        let b=FP4(Q.x)
        a.mul(Q.z); b.mul(z) 
        if !a.equals(b) {return false}
        a.copy(y); a.mul(Q.z)
        b.copy(Q.y); b.mul(z)
        if !a.equals(b) {return false}
    
        return true;
    }
    /* set self=-self */
    func neg()
    {
    //    if is_infinity() {return}
        y.norm(); y.neg(); y.norm()
        return
    }
    /* set to Affine - (x,y,z) to (x,y) */
    func affine() {
        if is_infinity() {return}
        let one=FP4(1)
        if z.equals(one) {
            x.reduce(); y.reduce()
            return
        }
        z.inverse()
    
        x.mul(z); x.reduce()
        y.mul(z); y.reduce()
        z.copy(one)
    }

    /* extract affine x as FP4 */
    func getX() -> FP4
    {
	let W=ECP4(); W.copy(self)
        W.affine()
        return W.x
    }
    /* extract affine y as FP4 */
    func getY() -> FP4
    {
	let W=ECP4(); W.copy(self)
        W.affine()
        return W.y
    }
    /* extract projective x */
    func getx() -> FP4
    {
        return x
    }
    /* extract projective y */
    func gety() -> FP4
    {
        return y
    }
    /* extract projective z */
    func getz() -> FP4
    {
        return z
    }

    /* convert to byte array */
    func toBytes(_ b:inout [UInt8])
    {
        let RM=Int(BIG.MODBYTES)
        var t=[UInt8](repeating: 0,count: RM)
	let W=ECP4(); W.copy(self)
        W.affine();
        
        W.x.geta().getA().toBytes(&t)
        for i in 0 ..< RM
            {b[i]=t[i]}
        W.x.geta().getB().toBytes(&t);
        for i in 0 ..< RM
            {b[i+RM]=t[i]}
    
        W.x.getb().getA().toBytes(&t)
        for i in 0 ..< RM
            {b[i+2*RM]=t[i]}
        W.x.getb().getB().toBytes(&t);
        for i in 0 ..< RM
            {b[i+3*RM]=t[i]}


        W.y.geta().getA().toBytes(&t);
        for i in 0 ..< RM
            {b[i+4*RM]=t[i]}
        W.y.geta().getB().toBytes(&t);
        for i in 0 ..< RM
            {b[i+5*RM]=t[i]}

        W.y.getb().getA().toBytes(&t);
        for i in 0 ..< RM
            {b[i+6*RM]=t[i]}
        W.y.getb().getB().toBytes(&t);
        for i in 0 ..< RM
            {b[i+7*RM]=t[i]}
      
    }

    /* convert from byte array to point */
    static func fromBytes(_ b:[UInt8]) -> ECP4
    {
        let RM=Int(BIG.MODBYTES)
        var t=[UInt8](repeating: 0,count: RM)

        for i in 0 ..< RM {t[i]=b[i]}
        let ra=BIG.fromBytes(t);
        for i in 0 ..< RM {t[i]=b[i+RM]}
        let rb=BIG.fromBytes(t);

        let ra2=FP2(ra,rb)

        for i in 0 ..< RM {t[i]=b[i+2*RM]}
        ra.copy(BIG.fromBytes(t));
        for i in 0 ..< RM {t[i]=b[i+3*RM]}
        rb.copy(BIG.fromBytes(t));

        let rb2=FP2(ra,rb)


        let rx=FP4(ra2,rb2)
    
        for i in 0 ..< RM {t[i]=b[i+4*RM]}
        ra.copy(BIG.fromBytes(t))
        for i in 0 ..< RM {t[i]=b[i+5*RM]}
        rb.copy(BIG.fromBytes(t))

        ra2.copy(FP2(ra,rb))

        for i in 0 ..< RM {t[i]=b[i+6*RM]}
        ra.copy(BIG.fromBytes(t))
        for i in 0 ..< RM {t[i]=b[i+7*RM]}
        rb.copy(BIG.fromBytes(t))

        rb2.copy(FP2(ra,rb))

        let ry=FP4(ra2,rb2)
    
        return ECP4(rx,ry)
    }

/* convert self to hex string */
    func toString() -> String
    {
	let W=ECP4(); W.copy(self)
        if W.is_infinity() {return "infinity"}
        W.affine()
        return "("+W.x.toString()+","+W.y.toString()+")"
    }
    
/* Calculate RHS of twisted curve equation x^3+B/i */
    static func RHS(_ x:FP4) -> FP4
    {
        x.norm()
        let r=FP4(x)
        r.sqr()
        let b=FP4(FP2(BIG(ROM.CURVE_B)))
        if ECP.SEXTIC_TWIST == ECP.D_TYPE {
            b.div_i()
        }
        if ECP.SEXTIC_TWIST == ECP.M_TYPE {
            b.times_i()
        }
        r.mul(x)
        r.add(b)
    
        r.reduce()
        return r
    }
/* construct self from (x,y) - but set to O if not on curve */
    public init(_ ix:FP4,_ iy:FP4)
    {
        x=FP4(ix)
        y=FP4(iy)
        z=FP4(1)
        let rhs=ECP4.RHS(x)
        let y2=FP4(y)
        y2.sqr()
        if !y2.equals(rhs) {inf()}
    }

    /* construct this from x - but set to O if not on curve */
    init(_ ix:FP4)
    {
        x=FP4(ix)
        y=FP4(1)
        z=FP4(1)
        let rhs=ECP4.RHS(x)
        if rhs.sqrt()
        {
            y.copy(rhs);
        }
        else {inf()}
    }

    /* this+=this */
    @discardableResult func dbl() -> Int
    {
    //    if (INF) {return -1}
        if y.iszilch()
        {
            inf();
            return -1;
        }
    
        let iy=FP4(y)
        if ECP.SEXTIC_TWIST == ECP.D_TYPE {       
            iy.times_i(); 
        }

        let t0=FP4(y) 
        t0.sqr();
        if ECP.SEXTIC_TWIST == ECP.D_TYPE {           
            t0.times_i() 
        }  
        let t1=FP4(iy)  
        t1.mul(z)
        let t2=FP4(z)
        t2.sqr()

        z.copy(t0)
        z.add(t0); z.norm() 
        z.add(z)
        z.add(z) 
        z.norm()  

        t2.imul(3*ROM.CURVE_B_I) 
        if ECP.SEXTIC_TWIST == ECP.M_TYPE {
            t2.times_i()  
        }
        let x3=FP4(t2)
        x3.mul(z) 

        let y3=FP4(t0)   

        y3.add(t2); y3.norm()
        z.mul(t1)
        t1.copy(t2); t1.add(t2); t2.add(t1); t2.norm()  
        t0.sub(t2); t0.norm()                           //y^2-9bz^2
        y3.mul(t0); y3.add(x3)                          //(y^2+3z*2)(y^2-9z^2)+3b.z^2.8y^2
        t1.copy(x); t1.mul(iy)                     //
        x.copy(t0); x.norm(); x.mul(t1); x.add(x)       //(y^2-9bz^2)xy2

        x.norm() 
        y.copy(y3); y.norm()
        return 1
    }

/* this+=Q - return 0 for add, 1 for double, -1 for O */
    @discardableResult func add(_ Q:ECP4) -> Int
    {

        let b=3*ROM.CURVE_B_I
        let t0=FP4(x)
        t0.mul(Q.x)         // x.Q.x
        let t1=FP4(y)
        t1.mul(Q.y)         // y.Q.y

        let t2=FP4(z)
        t2.mul(Q.z)
        let t3=FP4(x)
        t3.add(y); t3.norm()          //t3=X1+Y1
        let t4=FP4(Q.x)            
        t4.add(Q.y); t4.norm()         //t4=X2+Y2
        t3.mul(t4)                     //t3=(X1+Y1)(X2+Y2)
        t4.copy(t0); t4.add(t1)        //t4=X1.X2+Y1.Y2

        t3.sub(t4); t3.norm(); 
        if ECP.SEXTIC_TWIST == ECP.D_TYPE {
            t3.times_i()         //t3=(X1+Y1)(X2+Y2)-(X1.X2+Y1.Y2) = X1.Y2+X2.Y1
        }
        t4.copy(y)                    
        t4.add(z); t4.norm()           //t4=Y1+Z1
        let x3=FP4(Q.y)
        x3.add(Q.z); x3.norm()         //x3=Y2+Z2

        t4.mul(x3)                     //t4=(Y1+Z1)(Y2+Z2)
        x3.copy(t1)                    //
        x3.add(t2)                     //X3=Y1.Y2+Z1.Z2
    
        t4.sub(x3); t4.norm(); 
        if ECP.SEXTIC_TWIST == ECP.D_TYPE {  
            t4.times_i()          //t4=(Y1+Z1)(Y2+Z2) - (Y1.Y2+Z1.Z2) = Y1.Z2+Y2.Z1
        }
        x3.copy(x); x3.add(z); x3.norm()   // x3=X1+Z1
        let y3=FP4(Q.x)                
        y3.add(Q.z); y3.norm()             // y3=X2+Z2
        x3.mul(y3)                         // x3=(X1+Z1)(X2+Z2)
        y3.copy(t0)
        y3.add(t2)                         // y3=X1.X2+Z1+Z2
        y3.rsub(x3); y3.norm()             // y3=(X1+Z1)(X2+Z2) - (X1.X2+Z1.Z2) = X1.Z2+X2.Z1
        if ECP.SEXTIC_TWIST == ECP.D_TYPE {  
            t0.times_i() // x.Q.x
            t1.times_i() // y.Q.y
        }
        x3.copy(t0); x3.add(t0) 
        t0.add(x3); t0.norm()
        t2.imul(b)
        if ECP.SEXTIC_TWIST == ECP.M_TYPE {
            t2.times_i()
        }  
        let z3=FP4(t1); z3.add(t2); z3.norm()
        t1.sub(t2); t1.norm()
        y3.imul(b)
        if ECP.SEXTIC_TWIST == ECP.M_TYPE {          
            y3.times_i()
        }
        x3.copy(y3); x3.mul(t4); t2.copy(t3); t2.mul(t1); x3.rsub(t2)
        y3.mul(t0); t1.mul(z3); y3.add(t1)
        t0.mul(t3); z3.mul(t4); z3.add(t0)

        x.copy(x3); x.norm()
        y.copy(y3); y.norm()
        z.copy(z3); z.norm()    

        return 0
    }

    /* set self-=Q */
    @discardableResult func sub(_ Q:ECP4) -> Int
    {
	let NQ=ECP4(); NQ.copy(Q)
        NQ.neg()
        let D=add(NQ)
        //Q.neg()
        return D
    }


    static func frob_constants() -> [FP2]
    {
        let Fra=BIG(ROM.Fra)
        let Frb=BIG(ROM.Frb)
        let X=FP2(Fra,Frb)

        let f0=FP2(X); f0.sqr()
        let f2=FP2(f0)
        f2.mul_ip(); f2.norm()
        let f1=FP2(f2); f1.sqr()
        f2.mul(f1); f1.copy(X)
        if ECP.SEXTIC_TWIST == ECP.M_TYPE {
            f1.mul_ip()
            f1.inverse()
            f0.copy(f1); f0.sqr()

        }        
        f0.mul_ip(); f0.norm()
        f1.mul(f0)

        let F=[FP2(f0),FP2(f1),FP2(f2)];
        return F;
    }


/* set self*=q, where q is Modulus, using Frobenius */
    func frob(_ F:[FP2],_ n:Int)
    {
        for _ in 0 ..< n {
            x.frob(F[2])
            x.pmul(F[0])
        
            y.frob(F[2])
            y.pmul(F[1])
            y.times_i()

            z.frob(F[2])
        }
    }

    /* P*=e */
    func mul(_ e:BIG) -> ECP4
    {
    /* fixed size windows */
        let mt=BIG()
        let t=BIG()
        let P=ECP4()
        let Q=ECP4()
        let C=ECP4()
        
        var W=[ECP4]();
        for _ in 0 ..< 8 {W.append(ECP4())}
        
        var w=[Int8](repeating: 0,count: 1+(BIG.NLEN*Int(BIG.BASEBITS)+3)/4)
    
        if is_infinity() {return ECP4()}
    
        //affine()
    
    /* precompute table */
        Q.copy(self)
        Q.dbl()
        W[0].copy(self)
    
        for i in 1 ..< 8
        {
            W[i].copy(W[i-1])
            W[i].add(Q)
        }
    
    /* make exponent odd - add 2P if even, P if odd */
        t.copy(e)
        let s=t.parity()
        t.inc(1); t.norm(); let ns=t.parity(); mt.copy(t); mt.inc(1); mt.norm()
        t.cmove(mt,s)
        Q.cmove(self,ns)
        C.copy(Q)
    
        let nb=1+(t.nbits()+3)/4
    /* convert exponent to signed 4-bit window */
        for i in 0 ..< nb
        {
            w[i]=Int8(t.lastbits(5)-16)
            t.dec(Int(w[i])); t.norm()
            t.fshr(4)
        }
        w[nb]=Int8(t.lastbits(5))
    
        P.copy(W[Int(w[nb]-1)/2])
        for i in (0...nb-1).reversed()
        {
            Q.select(W,Int32(w[i]))
            P.dbl()
            P.dbl()
            P.dbl()
            P.dbl()
            P.add(Q)
        }
        P.sub(C);
        P.affine()
        return P;
    }
    
    /* P=u0.Q0+u1*Q1+u2*Q2+u3*Q3.. */
    // Bos & Costello https://eprint.iacr.org/2013/458.pdf
    // Faz-Hernandez & Longa & Sanchez  https://eprint.iacr.org/2013/158.pdf
    // Side channel attack secure 

    static func mul8(_ Q:[ECP4],_ u:[BIG]) -> ECP4
    {
        let W=ECP4()
        let P=ECP4()
        
        var T1=[ECP4]()
        var T2=[ECP4]()
               
        for _ in 0 ..< 8 {
            T1.append(ECP4())
            T2.append(ECP4())
        }
    
        let mt=BIG()
        var t=[BIG]()
    
        var w1=[Int8](repeating: 0,count: BIG.NLEN*Int(BIG.BASEBITS)+1)
        var s1=[Int8](repeating: 0,count: BIG.NLEN*Int(BIG.BASEBITS)+1)
    
        var w2=[Int8](repeating: 0,count: BIG.NLEN*Int(BIG.BASEBITS)+1)
        var s2=[Int8](repeating: 0,count: BIG.NLEN*Int(BIG.BASEBITS)+1)

        for i in 0 ..< 8
        {
            t.append(BIG(u[i]))
            t[i].norm()
            //Q[i].affine()
        }

    // precompute table 

        T1[0].copy(Q[0])  // Q[0]
        T1[1].copy(T1[0]); T1[1].add(Q[1])  // Q[0]+Q[1]
        T1[2].copy(T1[0]); T1[2].add(Q[2])  // Q[0]+Q[2]
        T1[3].copy(T1[1]); T1[3].add(Q[2])  // Q[0]+Q[1]+Q[2]
        T1[4].copy(T1[0]); T1[4].add(Q[3])  // Q[0]+Q[3]
        T1[5].copy(T1[1]); T1[5].add(Q[3])  // Q[0]+Q[1]+Q[3]
        T1[6].copy(T1[2]); T1[6].add(Q[3])  // Q[0]+Q[2]+Q[3]
        T1[7].copy(T1[3]); T1[7].add(Q[3])  // Q[0]+Q[1]+Q[2]+Q[3]

// Use Frobenius

        let F=ECP4.frob_constants()

        for i in 0 ..< 8 {
            T2[i].copy(T1[i]); T2[i].frob(F,4)
        }

// Make it odd
        let pb1=1-t[0].parity()
        t[0].inc(pb1)
        t[0].norm()  

        let pb2=1-t[4].parity()
        t[4].inc(pb2)
        t[4].norm()          

// Number of bits
        mt.zero();
        for i in 0 ..< 8 {
            mt.or(t[i]); 
        }

        let nb=1+mt.nbits()

// Sign pivot 

        s1[nb-1]=1
        s2[nb-1]=1
        for i in 0 ..< nb-1 {
            t[0].fshr(1)
            s1[i]=2*Int8(t[0].parity())-1
            t[4].fshr(1)
            s2[i]=2*Int8(t[4].parity())-1            
        }

// Recoded exponent
        for i in 0 ..< nb {
            w1[i]=0
            var k=1
            for j in 1 ..< 4 {
                let bt=s1[i]*Int8(t[j].parity())
                t[j].fshr(1)
                t[j].dec(Int(bt>>1))
                t[j].norm()
                w1[i]+=bt*Int8(k)
                k=2*k
            }
            w2[i]=0
            k=1
            for j in 5 ..< 8 {
                let bt=s2[i]*Int8(t[j].parity())
                t[j].fshr(1)
                t[j].dec(Int(bt>>1))
                t[j].norm()
                w2[i]+=bt*Int8(k)
                k=2*k
            }            
        }   

// Main loop
        P.select(T1,Int32(2*w1[nb-1]+1))
        W.select(T2,Int32(2*w2[nb-1]+1))
        P.add(W)
        for i in (0 ..< nb-1).reversed() {
            P.dbl()
            W.select(T1,Int32(2*w1[i]+s1[i]))
            P.add(W)
            W.select(T2,Int32(2*w2[i]+s2[i]))
            P.add(W)
        }    

        W.copy(P)  
        W.sub(Q[0])
        P.cmove(W,pb1) 

        W.copy(P)  
        W.sub(Q[4])
        P.cmove(W,pb2) 

        P.affine()
        return P
    }

     // needed for SOK
    static func mapit(_ h:[UInt8]) -> ECP4
    {
        let q=BIG(ROM.Modulus)
        var x=BIG.fromBytes(h)
        let one=BIG(1)
        var Q=ECP4()
        x.mod(q);
        while (true)
        {
            let X=FP4(FP2(one,x))
            Q=ECP4(X);
            if !Q.is_infinity() {break}
            x.inc(1); x.norm();
        }
    // Fast Hashing to G2 - Fuentes-Castaneda, Knapp and Rodriguez-Henriquez
        let F=ECP4.frob_constants()        
        x=BIG(ROM.CURVE_Bnx);
    
        let xQ=Q.mul(x);
        let x2Q=xQ.mul(x)
        let x3Q=x2Q.mul(x)
        let x4Q=x3Q.mul(x)

        if ECP.SIGN_OF_X == ECP.NEGATIVEX {
            xQ.neg()
            x3Q.neg()
        }

        x4Q.sub(x3Q)
        x4Q.sub(Q)

        x3Q.sub(x2Q)
        x3Q.frob(F,1)

        x2Q.sub(xQ)
        x2Q.frob(F,2)

        xQ.sub(Q)
        xQ.frob(F,3)

        Q.dbl()
        Q.frob(F,4)

        Q.add(x4Q)
        Q.add(x3Q)
        Q.add(x2Q)
        Q.add(xQ)
      
        Q.affine()
        return Q
    }  


    static public func generator() -> ECP4
    {
        return ECP4(FP4(FP2(BIG(ROM.CURVE_Pxaa),BIG(ROM.CURVE_Pxab)),FP2(BIG(ROM.CURVE_Pxba),BIG(ROM.CURVE_Pxbb))),FP4(FP2(BIG(ROM.CURVE_Pyaa),BIG(ROM.CURVE_Pyab)),FP2(BIG(ROM.CURVE_Pyba),BIG(ROM.CURVE_Pybb))))
    }

}


