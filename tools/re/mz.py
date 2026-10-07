import struct, math, sys, capstone, re
d=open(__import__('os').path.join(__import__('os').path.dirname(__file__),'..','..','PRIMA.EXE'),'rb').read()
hdr=struct.unpack_from('<14H',d,0)
cblp,cp,crlc,cparhdr=hdr[1],hdr[2],hdr[3],hdr[4]
ip,cs_=hdr[10],hdr[11]; lfarlc=hdr[12]
H=cparhdr*16; img=d[H:]
segs=set([cs_])
for k in range(crlc):
    o,s=struct.unpack_from('<HH',d,lfarlc+4*k)
    a=s*16+o
    if a+2<=len(img): segs.add(struct.unpack_from('<H',img,a)[0])
segs=sorted(x for x in segs if x*16<len(img))
def ext(b):
    m=int.from_bytes(b[:8],'little'); se=int.from_bytes(b[8:10],'little')
    s=-1 if se&0x8000 else 1; e=(se&0x7fff)-16383
    if not (m>>63): return 0.0 if m==0 else None
    return s*math.ldexp(m,e-63)
def segof(a):
    r=segs[0]
    for s in segs:
        if s*16<=a: r=s
    return r
if __name__=='__main__':
    print('header',hex(H),'entry',hex(cs_),hex(ip),'segs',[hex(s) for s in segs])
