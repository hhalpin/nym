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
//  pair192.swift
//
//  Created by Michael Scott on 07/07/2015.
//  Copyright (c) 2015 Michael Scott. All rights reserved.
//

/* AMCL BLS Curve Pairing functions */

final public class PAIR192 {
    
    // Line function
    static func line(_ A:ECP4,_ B:ECP4,_ Qx:FP,_ Qy:FP) -> FP24
    {
        var a:FP8
        var b:FP8
        var c:FP8

        if A===B
        { /* Doubling */

            let XX=FP4(A.getx())  //X
            let YY=FP4(A.gety())  //Y
            let ZZ=FP4(A.getz())  //Z
            let YZ=FP4(YY)        //Y 
            YZ.mul(ZZ)                //YZ
            XX.sqr()                  //X^2
            YY.sqr()                  //Y^2
            ZZ.sqr()                  //Z^2
            
            YZ.imul(4)
            YZ.neg(); YZ.norm()       //-2YZ
            YZ.qmul(Qy)               //-2YZ.Ys

            XX.imul(6)               //3X^2
            XX.qmul(Qx)              //3X^2.Xs

            let sb=3*ROM.CURVE_B_I
            ZZ.imul(sb)  
            if ECP.SEXTIC_TWIST == ECP.D_TYPE {             
                ZZ.div_2i();  
            }
            if ECP.SEXTIC_TWIST == ECP.M_TYPE {
                ZZ.times_i()
                ZZ.add(ZZ)
                YZ.times_i()
            }              
            ZZ.norm() // 3b.Z^2 

            YY.add(YY)
            ZZ.sub(YY); ZZ.norm()     // 3b.Z^2-Y^2

            a=FP8(YZ,ZZ)          // -2YZ.Ys | 3b.Z^2-Y^2 | 3X^2.Xs 
            if ECP.SEXTIC_TWIST == ECP.D_TYPE {             
                b=FP8(XX)            // L(0,1) | L(0,0) | L(1,0)
                c=FP8(0)
            } else { 
                b=FP8(0)
                c=FP8(XX); c.times_i()
            }        
            A.dbl()
        }
        else
        { // Addition
            let X1=FP4(A.getx())    // X1
            let Y1=FP4(A.gety())    // Y1
            let T1=FP4(A.getz())    // Z1
            let T2=FP4(A.getz())    // Z1
            
            T1.mul(B.gety())    // T1=Z1.Y2 
            T2.mul(B.getx())    // T2=Z1.X2

            X1.sub(T2); X1.norm()  // X1=X1-Z1.X2
            Y1.sub(T1); Y1.norm()  // Y1=Y1-Z1.Y2

            T1.copy(X1)            // T1=X1-Z1.X2
            X1.qmul(Qy)            // X1=(X1-Z1.X2).Ys
            if ECP.SEXTIC_TWIST == ECP.M_TYPE {
                X1.times_i()
            }              
            T1.mul(B.gety())       // T1=(X1-Z1.X2).Y2

            T2.copy(Y1)            // T2=Y1-Z1.Y2
            T2.mul(B.getx())       // T2=(Y1-Z1.Y2).X2
            T2.sub(T1); T2.norm()          // T2=(Y1-Z1.Y2).X2 - (X1-Z1.X2).Y2
            Y1.qmul(Qx);  Y1.neg(); Y1.norm() // Y1=-(Y1-Z1.Y2).Xs

            a=FP8(X1,T2)       // (X1-Z1.X2).Ys  |  (Y1-Z1.Y2).X2 - (X1-Z1.X2).Y2  | - (Y1-Z1.Y2).Xs
            if ECP.SEXTIC_TWIST == ECP.D_TYPE {              
                b=FP8(Y1)
                c=FP8(0)
            } else {
                b=FP8(0)
                c=FP8(Y1); c.times_i()
            }  
            A.add(B)
        }
        return FP24(a,b,c)
    }

    // Optimal R-ate pairing
    static public func ate(_ P1:ECP4,_ Q1:ECP) -> FP24
    {
        let x=BIG(ROM.CURVE_Bnx)
        let n=BIG(x)
        
        var lv:FP24

        let n3=BIG(n)
        n3.pmul(3)
        n3.norm()

	let P=ECP4(); P.copy(P1); P.affine()
	let Q=ECP(); Q.copy(Q1); Q.affine()


        let Qx=FP(Q.getx())
        let Qy=FP(Q.gety())
    
        let A=ECP4()
        let r=FP24(1)
    
        A.copy(P)
	let NP=ECP4()
	NP.copy(P)
	NP.neg()

        let nb=n3.nbits()
    
        for i in (1...nb-2).reversed()
        //for var i=nb-2;i>=1;i--
        {
            r.sqr()            
            lv=line(A,A,Qx,Qy)
            r.smul(lv,ECP.SEXTIC_TWIST)
            let bt=n3.bit(UInt(i))-n.bit(UInt(i))
            if bt == 1 {
              lv=line(A,P,Qx,Qy)
              r.smul(lv,ECP.SEXTIC_TWIST)
            }
            if bt == -1 {
                //P.neg()
                lv=line(A,NP,Qx,Qy)
                r.smul(lv,ECP.SEXTIC_TWIST)
                //P.neg()
            }
        }
    
        if ECP.SIGN_OF_X == ECP.NEGATIVEX {
            r.conj()
         }     

        return r
    }    

    // Optimal R-ate double pairing e(P,Q).e(R,S)
    static public func ate2(_ P1:ECP4,_ Q1:ECP,_ R1:ECP4,_ S1:ECP) -> FP24
    {
        let x=BIG(ROM.CURVE_Bnx)
        let n=BIG(x)
        var lv:FP24

        let n3=BIG(n)
        n3.pmul(3)
        n3.norm()
    
	let P=ECP4(); P.copy(P1); P.affine()
	let Q=ECP(); Q.copy(Q1); Q.affine()
	let R=ECP4(); R.copy(R1); R.affine()
	let S=ECP(); S.copy(S1); S.affine()


        let Qx=FP(Q.getx())
        let Qy=FP(Q.gety())
        let Sx=FP(S.getx())
        let Sy=FP(S.gety())
    
        let A=ECP4()
        let B=ECP4()
        let r=FP24(1)
    
        A.copy(P)
        B.copy(R)
	let NP=ECP4()
	NP.copy(P)
	NP.neg()
	let NR=ECP4()
	NR.copy(R)
	NR.neg()


        let nb=n3.nbits()
    
        for i in (1...nb-2).reversed()
        {
            r.sqr()            
            lv=line(A,A,Qx,Qy)
            r.smul(lv,ECP.SEXTIC_TWIST)
            lv=line(B,B,Sx,Sy)
            r.smul(lv,ECP.SEXTIC_TWIST)
            let bt=n3.bit(UInt(i))-n.bit(UInt(i))

            if bt == 1 {
                lv=line(A,P,Qx,Qy)
                r.smul(lv,ECP.SEXTIC_TWIST)
                lv=line(B,R,Sx,Sy)
                r.smul(lv,ECP.SEXTIC_TWIST)
            }

            if bt == -1 {
                //P.neg(); 
                lv=line(A,NP,Qx,Qy)
                r.smul(lv,ECP.SEXTIC_TWIST)
                //P.neg(); 
                //R.neg()
                lv=line(B,NR,Sx,Sy)
                r.smul(lv,ECP.SEXTIC_TWIST)
                //R.neg()                
            }            

        }
    
        if ECP.SIGN_OF_X == ECP.NEGATIVEX {
            r.conj()
        }     

        return r
    }
    
    // final exponentiation - keep separate for multi-pairings and to avoid thrashing stack
    static public func fexp(_ m:FP24) -> FP24
    {
        let f=FP2(BIG(ROM.Fra),BIG(ROM.Frb));
        let x=BIG(ROM.CURVE_Bnx)
        let r=FP24(m)
    
    // Easy part of final exp
        let lv=FP24(r)
        lv.inverse()
        r.conj()
    
        r.mul(lv)
        lv.copy(r)
        r.frob(f,4)
        r.mul(lv)
        
    // Hard part of final exp

        let t7=FP24(r); t7.usqr()
        let t1=t7.pow(x)

        x.fshr(1)
        let t2=t1.pow(x)
        x.fshl(1)

        if ECP.SIGN_OF_X==ECP.NEGATIVEX {
            t1.conj()
        }
        let t3=FP24(t1); t3.conj()
        t2.mul(t3)
        t2.mul(r)

        t3.copy(t2.pow(x))
        let t4=t3.pow(x)
        let t5=t4.pow(x)

        if ECP.SIGN_OF_X==ECP.NEGATIVEX {
            t3.conj(); t5.conj()
        }

        t3.frob(f,6); t4.frob(f,5)
        t3.mul(t4);

        let t6=t5.pow(x)
        if ECP.SIGN_OF_X==ECP.NEGATIVEX {
            t6.conj()
        }

        t5.frob(f,4)
        t3.mul(t5)

        let t0=FP24(t2); t0.conj()
        t6.mul(t0)

        t5.copy(t6)
        t5.frob(f,3)

        t3.mul(t5)
        t5.copy(t6.pow(x))
        t6.copy(t5.pow(x))

        if ECP.SIGN_OF_X==ECP.NEGATIVEX {
            t5.conj()
        }

        t0.copy(t5)
        t0.frob(f,2)
        t3.mul(t0)
        t0.copy(t6)
        t0.frob(f,1)

        t3.mul(t0)
        t5.copy(t6.pow(x))

        if ECP.SIGN_OF_X==ECP.NEGATIVEX {
            t5.conj()
        }
        t2.frob(f,7)

        t5.mul(t7)
        t3.mul(t2)
        t3.mul(t5)

        r.mul(t3)

        r.reduce()
        return r
    }

    // GLV method
    static func glv(_ e:BIG) -> [BIG]
    {
        var u=[BIG]();
        let q=BIG(ROM.CURVE_Order)
        let x=BIG(ROM.CURVE_Bnx)
        let x2=BIG.smul(x,x)
        x.copy(BIG.smul(x2,x2))
        u.append(BIG(e))
        u[0].mod(x)
        u.append(BIG(e))
        u[1].div(x)
        u[1].rsub(q)

        return u
    }

    // Galbraith & Scott Method
    static func gs(_ e:BIG) -> [BIG]
    {
        var u=[BIG]();
        let q=BIG(ROM.CURVE_Order)        
        let x=BIG(ROM.CURVE_Bnx)
        let w=BIG(e)
        for i in 0 ..< 7
        {
            u.append(BIG(w))
            u[i].mod(x)
            w.div(x)
        }
        u.append(BIG(w))
        if ECP.SIGN_OF_X == ECP.NEGATIVEX {
            u[1].copy(BIG.modneg(u[1],q))
            u[3].copy(BIG.modneg(u[3],q))         
            u[5].copy(BIG.modneg(u[5],q))
            u[7].copy(BIG.modneg(u[7],q))                       
        }        

        return u
    }   


    // Multiply P by e in group G1
    static public func G1mul(_ P:ECP,_ e:BIG) -> ECP
    {
        var R:ECP
        if (ROM.USE_GLV)
        {
            //P.affine()
            R=ECP()
            R.copy(P)
            let Q=ECP()
            Q.copy(P); Q.affine()
            let q=BIG(ROM.CURVE_Order)
            let cru=FP(BIG(ROM.CURVE_Cru))
            let t=BIG(0)
            var u=PAIR192.glv(e)
            Q.getx().mul(cru);
    
            var np=u[0].nbits()
            t.copy(BIG.modneg(u[0],q))
            var nn=t.nbits()
            if (nn<np)
            {
                u[0].copy(t)
                R.neg()
            }
    
            np=u[1].nbits()
            t.copy(BIG.modneg(u[1],q))
            nn=t.nbits()
            if (nn<np)
            {
                u[1].copy(t)
                Q.neg()
            }
            u[0].norm()
            u[1].norm()
            R=R.mul2(u[0],Q,u[1])
        }
        else
        {
            R=P.mul(e)
        }
        return R
    }

    // Multiply P by e in group G2
    static public func G2mul(_ P:ECP4,_ e:BIG) -> ECP4
    {
        var R:ECP4
        if (ROM.USE_GS_G2)
        {
            var Q=[ECP4]()
            let F=ECP4.frob_constants()
            let q=BIG(ROM.CURVE_Order);
            var u=PAIR192.gs(e);
    
            let t=BIG(0)
            //P.affine()
            Q.append(ECP4())
            Q[0].copy(P);
            for i in 1 ..< 8
            {
                Q.append(ECP4()); Q[i].copy(Q[i-1]);
                Q[i].frob(F,1);
            }
            for i in 0 ..< 8
            {
                let np=u[i].nbits()
                t.copy(BIG.modneg(u[i],q))
                let nn=t.nbits()
                if (nn<np)
                {
                    u[i].copy(t)
                    Q[i].neg()
                }
                u[i].norm()
            }
    
            R=ECP4.mul8(Q,u)
        }
        else
        {
            R=P.mul(e)
        }
        return R;
    }

    // f=f^e
    // Note that this method requires a lot of RAM! Better to use compressed XTR method, see FP8.swift
    static public func GTpow(_ d:FP24,_ e:BIG) -> FP24
    {
        var r:FP24
        if (ROM.USE_GS_GT)
        {
            var g=[FP24]()
            let f=FP2(BIG(ROM.Fra),BIG(ROM.Frb))
            let q=BIG(ROM.CURVE_Order)
            let t=BIG(0)
        
            var u=gs(e)
            g.append(FP24(0))
            g[0].copy(d);
            for i in 1 ..< 8
            {
                g.append(FP24(0)); g[i].copy(g[i-1])
                g[i].frob(f,1)
            }
            for i in 0 ..< 8
            {
                let np=u[i].nbits()
                t.copy(BIG.modneg(u[i],q))
                let nn=t.nbits()
                if (nn<np)
                {
                    u[i].copy(t)
                    g[i].conj()
                }
                u[i].norm()                
            }
            r=FP24.pow8(g,u)
        }
        else
        {
            r=d.pow(e)
        }
        return r
    }
}
