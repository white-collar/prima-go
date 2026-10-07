; FEEDER.COM - TSR: serves keystrokes from KEYS.BIN via INT 16h.
; KEYS.BIN = sequence of words: (scan<<8)|ascii; 0FFFFh = dump screen to SCRnn.BIN
        org 100h
start:  jmp install
old16   dd 0
pos     dw 0
len     dw 0
count   db 0
fname   db 'SCR00.BIN',0
kname   db 'KEYS.BIN',0

; process pending dump markers; returns ZF=1 if script exhausted, else AX=next key
peek:   push si
.again: mov si,[cs:pos]
        cmp si,[cs:len]
        jae .empty
        mov ax,[cs:buf+si]
        cmp ax,0FFFFh
        jne .have
        call dump
        add word [cs:pos],2
        jmp .again
.have:  pop si
        or si,si        ; dummy to set ZF=0 reliably below
        cmp ax,0FFFFh   ; ax != FFFF so ZF=0
        ret
.empty: pop si
        cmp ax,ax       ; ZF=1
        ret

dump:   push ax
        push bx
        push cx
        push dx
        push ds
        push cs
        pop ds
        mov al,[count]
        aam
        add ax,3030h
        mov [fname+3],ah
        mov [fname+4],al
        inc byte [count]
        mov ah,3Ch
        xor cx,cx
        mov dx,fname
        int 21h
        jc .out
        mov bx,ax
        mov ax,0B800h
        mov ds,ax
        xor dx,dx
        mov cx,4000
        mov ah,40h
        int 21h
        mov ah,3Eh
        int 21h
.out:   pop ds
        pop dx
        pop cx
        pop bx
        pop ax
        ret

int16:  cmp ah,0
        je .read
        cmp ah,10h
        je .read
        cmp ah,1
        je .check
        cmp ah,11h
        je .check
.chain: jmp far [cs:old16]
.read:  push ax
        call peek
        jz .rnone
        add word [cs:pos],2
        add sp,2
        iret
.rnone: pop ax
        jmp .chain
.check: push ax
        call peek
        jz .rnone
        add sp,2
        push bp
        mov bp,sp
        and word [bp+6],0FFBFh   ; clear ZF in caller flags => key available
        pop bp
        iret

buf:    times 8192 db 0

install:
        mov ax,3D00h
        mov dx,kname
        int 21h
        jc .nofile
        mov bx,ax
        mov ah,3Fh
        mov cx,8192
        mov dx,buf
        int 21h
        mov [len],ax
        mov ah,3Eh
        int 21h
.nofile:
        mov ax,3516h
        int 21h
        mov [old16],bx
        mov [old16+2],es
        mov ax,2516h
        mov dx,int16
        int 21h
        mov dx,(install-start+100h+15)/16
        mov ax,3100h
        int 21h
