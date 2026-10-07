import sys, re, struct
import mz, fix
from capstone import *
img=mz.img; segs=mz.segs
md=Cs(CS_ARCH_X86,CS_MODE_16)
# Entry points in the Borland System unit (segment 2b04)
RTL={0xc88:'sqrt',0xc8c:'sin',0xc91:'cos',0xc96:'arctan',0x530:'STACKCHK',0xc6c:'frac',0xc5c:'int'}
def ext(b): return mz.ext(b)
def fmt(v):
    if v is None: return '?'
    return ('%.12g'%v)
def P(x): return x if re.fullmatch(r'[\w.\[\]:*@$]+|-?[\d.]+(e-?\d+)?',x) else '('+x+')'
FULL=False
def run(seg,start,stop=None,quiet=False,localnames={}):
    base=seg*16; nxt=[s*16 for s in segs if s>seg]; end=nxt[0] if nxt else len(img)
    code=img[base:end]
    st=[]; out=[]; esname='es:?'
    def T(k=0): return st[-1-k] if len(st)>k else f'?{k}'
    def S(k,v):
        while len(st)<=k: st.insert(0,'?')
        st[-1-k]=v
    def memname(ins):
        op=ins.op_str
        m=re.search(r'cs:\[(0x[0-9a-f]+)\]',op)
        if m:
            a=int(m.group(1),16)
            if 'xword' in op or 'tbyte' in op: return fmt(ext(code[a:a+10]))
            if 'qword' in op: return fmt(struct.unpack_from('<d',code,a)[0])
            if 'dword' in op: return fmt(struct.unpack_from('<f',code,a)[0])
            return f'cs[{a:x}]'
        m=re.search(r'\[bp ([+-]) (0x[0-9a-f]+|\d+)\]',op)
        if m:
            n=int(m.group(2),0); key=('p' if m.group(1)=='+' else 'v')+f'{n:x}'
            return localnames.get(key,key)
        if 'es:[di' in op:
            inner=re.search(r'es:\[di( \+ (0x[0-9a-f]+|\d+))?\]',op)
            off=int(inner.group(2),0) if inner and inner.group(2) else 0
            return esname+(f'+{off:x}' if off else '')
        m=re.search(r'\[(0x[0-9a-f]+)\]',op)
        if m: return f'g{int(m.group(1),16):x}'
        return op
    off=start
    for ins in md.disasm(code[start:end-base if stop is None else stop],start):
        b=ins.bytes; mn=ins.mnemonic; a=ins.address
        if a!=start and mn=='push' and ins.op_str=='bp' and stop is None: break
        if mn=='les' and 'bp' in ins.op_str:
            m=re.search(r'\[bp ([+-]) (0x[0-9a-f]+|\d+)\]',ins.op_str)
            if m:
                n=int(m.group(2),0); key=('p' if m.group(1)=='+' else 'v')+f'{n:x}'
                esname='*'+localnames.get(key,key)
            continue
        if mn=='mov' and ins.op_str.startswith('di, sp'): esname='ARG'; continue
        # strip fwait prefix
        bb=bytes(b)
        while bb and bb[0] in (0x9b,): bb=bb[1:]
        if bb and bb[0] in (0x26,0x2e,0x36,0x3e): pre=bb[0]; bb=bb[1:]
        if mn in('call','lcall'):
            ops=ins.op_str.split(', ')
            if mn=='lcall' and int(ops[0],16)==0x2b04:
                f=RTL.get(int(ops[1],16),'rtl_'+ops[1])
                if f=='STACKCHK': continue
                if f in('sqrt','sin','cos','arctan','frac','int'): S(0,f'{f}({T()})')
                else: out.append(f'{a:04x}  call {f}   st0={T()}')
            else:
                tgt=ops[-1]; out.append(f'{a:04x}  CALL {tgt}   (st0={T()})')
                st.append(f'F{tgt}()')
            continue
        if mn in('ret','retf'):
            out.append(f'{a:04x}  RET  st0={T()}')
            if stop is None and not FULL: break
            continue
        if mn.startswith('j'): out.append(f'{a:04x}  {mn} {ins.op_str}'); continue
        if not mn.startswith('f') or len(bb)<2: continue
        o,r=bb[0],bb[1]
        if r>=0xc0 and 0xd8<=o<=0xdf:   # register form
            i=r&7; g=(r>>3)&7
            if o==0xd9:
                if g==0: st.append(T(i))
                elif g==1: x,y=T(),T(i); S(0,y); S(i,x)
                elif r==0xe0: S(0,'-'+P(T()))
                elif r==0xe1: S(0,f'abs({T()})')
                elif r==0xe4: out.append(f'{a:04x}  test {T()}')
                elif r==0xe8: st.append('1')
                elif r==0xee: st.append('0')
                elif r==0xeb: st.append('pi')
                elif r==0xfa: S(0,f'sqrt({T()})')
                elif r==0xfc: S(0,f'round({T()})')
                else: out.append(f'{a:04x}  ?? {mn} {ins.op_str}')
                continue
            if o==0xdd and g in(2,3):
                S(i,T());
                if g==3 and st: st.pop()
                continue
            if o==0xde and r==0xd9: out.append(f'{a:04x}  compare {T()} ? {T(1)}'); st[-2:]=[]; continue
            if o==0xd8 and g in(2,3):
                out.append(f'{a:04x}  compare {T()} ? {T(i)}')
                if g==3 and st: st.pop()
                continue
            if o in(0xd8,0xdc,0xde):
                ops={0:'+',1:'*',4:'-',5:'-r',6:'/',7:'/r'}
                if g not in ops: out.append(f'{a:04x}  ?? {mn}'); continue
                if o==0xd8:
                    x,y=T(),T(i); op=ops[g]
                    if op.endswith('r'): x,y=y,x
                    S(0,f'{P(x)} {op[0]} {P(y)}')
                else:
                    # dc/de: dest st(i). Intel: E0=fsubr(st0-sti) E8=fsub(sti-st0) F0=fdivr(st0/sti) F8=fdiv(sti/st0)
                    x,y=T(i),T()
                    op={0:'+',1:'*',4:'-r',5:'-',6:'/r',7:'/'}[g]
                    if op.endswith('r'): x,y=y,x
                    S(i,f'{P(x)} {op[0]} {P(y)}')
                    if o==0xde and st: st.pop()
                continue
            out.append(f'{a:04x}  ?? {mn} {ins.op_str}'); continue
        # memory forms
        g=(r>>3)&7; M=memname(ins)
        if o in(0xd8,0xdc,0xda,0xde):
            if g in(2,3):
                out.append(f'{a:04x}  compare {T()} ? {M}'); 
                if g==3 and st: st.pop()
                continue
            op={0:'+',1:'*',4:'-',5:'-r',6:'/',7:'/r'}[g]
            x,y=T(),M
            if op.endswith('r'): x,y=y,x
            S(0,f'{P(x)} {op[0]} {P(y)}'); continue
        if (o in(0xd9,0xdd) and g==0) or (o==0xdb and g in(0,5)) or (o in(0xdf,) and g in(0,5)): st.append(M); continue
        if (o in(0xd9,0xdd,0xdb,0xdf) and g in(2,3,7)) or (o==0xdb and g==7):
            if g in (2,3,7):
                out.append(f'{a:04x}  {M} := {T()}')
                if g in(3,7): (st.pop() if st else None)
                continue
        if o in(0xd9,0xdd) and g in(5,7,6,4): continue  # fldcw/fstcw/fstsw etc
        out.append(f'{a:04x}  ?? {mn} {ins.op_str}')
    return out
if __name__=='__main__':
    seg=int(sys.argv[1],16); start=int(sys.argv[2],16)
    names=dict(kv.split('=') for kv in sys.argv[3:])
    print('\n'.join(run(seg,start,localnames=names)))
