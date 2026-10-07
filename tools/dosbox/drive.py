"""Drive PRIMA.EXE in DOSBox-X: drive.py "<tokens>"  -> prints screen dumps.
Tokens: single chars, or names: enter esc tab down up left right home end bs f1..f10 sp, and '#' = dump screen."""
import os, subprocess, sys, time, glob, struct
S=os.path.dirname(os.path.abspath(__file__)); D=os.path.join(S,'work')
def setup():
    import shutil
    os.makedirs(D,exist_ok=True)
    shutil.copy(os.path.join(S,'..','..','PRIMA.EXE'),D)
    subprocess.run(['nasm','-f','bin','-o',os.path.join(D,'FEEDER.COM'),os.path.join(S,'feeder.asm')],check=True)
NAMED={'enter':0x1C0D,'esc':0x011B,'tab':0x0F09,'down':0x5000,'up':0x4800,'left':0x4B00,'right':0x4D00,
       'home':0x4700,'end':0x4F00,'bs':0x0E08,'sp':0x3920,'del':0x5300,'ins':0x5200,'pgdn':0x5100,'pgup':0x4900,'#':0xFFFF}
for i in range(1,11): NAMED[f'f{i}']=(0x3A+i)<<8
ROW='1234567890-='; QW='qwertyuiop'; AS='asdfghjkl'; ZX='zxcvbnm,./'
def code(ch):
    if ch in ROW: return ((2+ROW.index(ch))<<8)|ord(ch)
    l=ch.lower()
    for base,row in((0x10,QW),(0x1E,AS),(0x2C,ZX)):
        if l in row: return ((base+row.index(l))<<8)|ord(ch)
    return ord(ch)
def build(tokens):
    out=[]
    for t in tokens.split():
        if t in NAMED: out.append(NAMED[t])
        else: out+= [code(c) for c in t]
    return out
def run(tokens, timeout=40):
    setup()
    keys=build(tokens); ndump=keys.count(0xFFFF)
    for f in glob.glob(D+'/SCR*.BIN'): os.remove(f)
    open(D+'/KEYS.BIN','wb').write(struct.pack(f'<{len(keys)}H',*keys))
    conf=os.path.join(D,'drive.conf')
    open(conf,'w').write(f"""[sdl]
output=surface
[dosbox]
machine=svga_s3
[cpu]
cycles=max
[autoexec]
mount c "{D}"
c:
FEEDER.COM
PRIMA.EXE
""")
    p=subprocess.Popen(['dosbox-x','-conf',conf,'-nopromptfolder','-fastlaunch'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
    t0=time.time()
    while time.time()-t0<timeout and len(glob.glob(D+'/SCR*.BIN'))<ndump: time.sleep(0.5)
    time.sleep(1.5); p.kill(); p.wait()
    screens=[]
    for f in sorted(glob.glob(D+'/SCR*.BIN')):
        d=open(f,'rb').read()
        screens.append([bytes(d[2*(r*80+c)] for c in range(80)).decode('cp866','replace').rstrip() for r in range(25)])
    return screens
if __name__=='__main__':
    for i,s in enumerate(run(sys.argv[1])):
        print(f'=== screen {i}'); print('\n'.join(l for l in s if l.strip('║ ')))
