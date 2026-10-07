"""Run every case in cases.py through PRIMA.EXE (DOSBox-X) and the Go port, and compare."""
import json, re, subprocess, os, cases, drive
H=os.path.dirname(os.path.abspath(__file__))
orig={}
for name,(path,rows) in cases.CASES.items():
    keys='enter t1 enter '+path+' '+' '.join(cases.tokens(r) for r in rows)+' #'
    orig[name]=drive.run(keys)[-1]
inp=json.dumps({k:[list(r) for r in v[1]] for k,v in cases.CASES.items()})
go=json.loads(subprocess.run(['go','run','./tools/dosbox/gocalc'],input=inp,capture_output=True,text=True,check=True,cwd=os.path.join(H,'..','..')).stdout)
# name: (outputs per row, which cells, is_angle flags, two-line row?)
SPEC={'radii':(slice(2,6),[0,0,0,0],False),'RA':(slice(3,4),[0],False),'merid':(slice(3,4),[0],False),
'paral':(slice(3,4),[0],False),'area':(slice(5,6),[0],False),'bl2xy':(slice(4,6),[0,0],False),
'gamma':(slice(4,5),[1],False),'xy2bl':(slice(4,6),[1,1],False),
'schreib':(slice(1,4),[1,1,1],True),'rkm':(slice(1,4),[1,1,1],True),'inverse':(slice(1,5),[1,1,1,0],True),
'corr':(slice(1,4),[0,0,0],True),'frames':(slice(4,7),[0,0,0],True)}
def num(cell,isang):
    cell=cell.strip()
    if not isang: return float(cell)
    neg=cell.startswith('-'); d,m,s=cell.lstrip('-').split()
    v=float(d)*3600+float(m)*60+float(s); return -v if neg else v
worst={}
for name,(sl,angs,two) in SPEC.items():
    lines=[l for l in orig[name] if l.startswith('│')]
    rows=[]
    for i,l in enumerate(lines):
        c=l.split('│')[1:-1]
        if re.fullmatch(r'\s*\d+\s*',c[0]) and (not two or i+1<len(lines)):
            src=lines[i+1].split('│')[1:-1] if two else c
            try: rows.append([num(x,a) for x,a in zip(src[sl],angs)])
            except Exception: pass
    print(f'--- {name}')
    for k,(o,gv) in enumerate(zip(rows,go[name])):
        diffs=[abs(a-b) for a,b in zip(o,gv)]
        print(f'  row {k+1}: orig={o}\n         go  ={[round(x,4) for x in gv]}\n         |diff|={[round(x,4) for x in diffs]}')
