.origin 0
.entrypoint START

#include "spislave.hp"

START:
    MOV r8, 255
    SBCO r8, CONST_PRUDRAM, 6, 1

    // Enable OCP master port
    LBCO r0, CONST_PRUCFG, 4, 4
    CLR  r0, r0, 4
    SBCO r0, CONST_PRUCFG, 4, 4

    //C28 will point to 0x00012000 (PRU shared RAM)
    MOV  r0, 0x00000120
    MOV  r1, CTPPR_0
    ST32 r0, r1
  
    // Enable CLKSPIREF and CLK
    MOV  r1, 0x44E00050
    MOV  r2,  0x00000002
    SBBO r2,  r1, 0, 4
    
    // Reset SPI
    MOV  r1, 0x481A0110
    LBBO r2,  r1, 0, 4
    SET  r2.t1
    SBBO r2,  r1, 0, 4
    
    //Wait for RESET
    RESET:
      MOV  r1,  0x481a0114
      LBBO r2,   r1, 0, 4
      QBBC RESET, r2.t0
    
    //Config MODULCTRL
    MOV  r1, 0x481A0128
    MOV  r2,  0x00000000
    SBBO r2,  r1 , 0, 4
    
    //Config SYSCONFIG
    MOV  r1, 0x481A0110
    MOV  r2,  0x00000311
    SBBO r2,  r1, 0, 4
    
    //Reset interrupt status bits
    MOV  r1, 0x481A0118
    MOV  r2,  0xFFFFFFFF
    SBBO r2,  r1, 0, 4
    
    //Disable interupts
    MOV  r1, 0x481A011C
    MOV  r2,  0x00000000
    SBBO r2,  r1, 0, 4
  
    // Disable channel 1
    MOV  r1, 0x481A0148
    MOV  r2,  0x00000000
    SBBO r2,  r1, 0, 4
    
    // Configure channel 1 of MCSPI1
    MOV  r1, 0x481A0140  
    MOV  r2, 0x000192FD0
//    LBCO r4, CONST_PRUDRAM, 0, 4 //frequency
//    OR   r2, r2, r4
    SBBO r2, r1, 0, 4

// setup complete
MOV r8, 254
SBCO r8, CONST_PRUDRAM, 6, 1

LOOPING:
    // Exit this program if the exit flag is cleared
    LBCO r2, CONST_PRUDRAM, 4, 1
    QBEQ EXIT, r2.b0, 0

    // Fast loop if the ready flag is not set
    LBCO r2, CONST_PRUDRAM, 5, 1
    QBEQ LOOPING, r2.b0, 0

    MOV r8, 253
    SBCO r8, CONST_PRUDRAM, 6, 1

    // Data is ready, read data length word
    //LBCO r5, CONST_PRUDRAM, 7, 4
    MOV r5, 5
    // Now we read a byte from PRUDRAM and write it to spi

START_FRAME:
    // First send 4 zero bytes of startframe
    MOV r1, 0x481A014C
    MOV r2, 0
    SBBO r2, r1, 0, 4

    // send 4 zero bytes
    CALL ENABLE_CH1

    MOV r9, 11

DATA_FRAME:
    // now send data bytes 4 at a time, which is convenient
    // since its the length of an SPI word
    // first we need to read 3 bytes from the buffer at current offset
    LBCO r2, CONST_PRUDRAM, r9, 3
    LSR r2, r2, 8
    MOV r3, 0xE4000000
    OR r2, r2, r3
    // r2 is now ready to send
    MOV r1, 0x481A014C
    SBBO r2, r1, 0, 4
    CALL ENABLE_CH1
    // increment r9 offset and decrement r5 counter
    SUB r5, r5, 1
    ADD r9, r9, 3
    QBNE DATA_FRAME, r5, 0
    

END_FRAME:
    // now send end frame bytes, 
    MOV r6, 0
    //LBCO r5, CONST_PRUDRAM, 7, 4
    MOV r5, 5
 END_FRAME_LOOP:
    MOV r1, 0x481A014C
    MOV r2, 0xFFFFFFFF
    SBBO r2, r1, 0, 4
    CALL ENABLE_CH1
    ADD r6, r6, 64
    QBLT END_FRAME_LOOP, r5, r6

// reset ready flag
MOV r2, 0
SBCO r2, CONST_PRUDRAM, 5, 1
QBA LOOPING

CHECKTX1:
MOV  r1, 0x481A0144
LBBO r2, r1, 0, 4
QBBC CHECKTX1, r2.t1
JMP  r19.w0

ENABLE_CH1:
// Enable Channel 1
MOV  r1, 0x481A0148
MOV  r2, 0x00000001
SBBO r2, r1, 0, 4
JAL  r19.w0, CHECKTX1
RET

EXIT:

MOV r8, 250
SBCO r8, CONST_PRUDRAM, 6, 1

#ifdef AM33XX
    // Send notification to Host for program completion
    MOV R31.b0, PRU0_ARM_INTERRUPT+16
#else
    MOV R31.b0, PRU0_ARM_INTERRUPT
#endif

    HALT
