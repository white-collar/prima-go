# patch Borland FPU-emulator interrupts back to real 8087 opcodes
import mz
b=bytearray(mz.img); i=0
segp={0x00:0x3e,0x40:0x36,0x80:0x2e,0xc0:0x26}
while i<len(b)-2:
    if b[i]==0xcd and 0x34<=b[i+1]<=0x3b: b[i]=0x9b; b[i+1]=0xd8+b[i+1]-0x34; i+=2; continue
    if b[i]==0xcd and b[i+1]==0x3c:
        x=b[i+2]; b[i]=0x9b; b[i+1]=segp[x&0xc0]; b[i+2]=(x&0x3f)|0xc0; i+=3; continue
    if b[i]==0xcd and b[i+1]==0x3d: b[i]=0x9b; b[i+1]=0x90; i+=2; continue
    i+=1
mz.img=bytes(b)
