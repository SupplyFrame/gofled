#include "spislave.h"

#include <stdio.h>
#include <prussdrv.h>
#include <pruss_intc_mapping.h>

#define PRU_NUM 0
#define AM33XX

static void *pruDataMem;
static unsigned char *pruDataMem_byte;

unsigned char* spiinit() {
	unsigned int ret;
	tpruss_intc_initdata pruss_intc_initdata = PRUSS_INTC_INITDATA;

	prussdrv_init();
	ret = prussdrv_open(PRU_EVTOUT_0);

	if (ret) {
		printf("prussdrv_open open failed\n");
		return;
	}

	prussdrv_pruintc_init(&pruss_intc_initdata);

	printf("\tINFO: Initializing example.\r\n");
	// LOCAL_exampleinit();
	prussdrv_map_prumem(PRUSS0_PRU0_DATARAM, &pruDataMem);
	pruDataMem_byte = (unsigned char*) pruDataMem;

	pruDataMem_byte[0] = 0; // 0x00000010 = 3000000
	pruDataMem_byte[1] = 0;
	pruDataMem_byte[2] = 0;
	pruDataMem_byte[3] = 0x10;

	// exit flag
	pruDataMem_byte[4] = 1;
	// ready flag
	pruDataMem_byte[5] = 0;
	// feedback flag, read this to know state of PRU
	pruDataMem_byte[6] = 0;
	// data length
	pruDataMem_byte[7] = 0;
	pruDataMem_byte[8] = 0;
	pruDataMem_byte[9] = 0;
	pruDataMem_byte[10] = 1;

	prussdrv_exec_program(PRU_NUM, "./spislave.bin");

	return pruDataMem_byte;
}

void spiclose() {
	pruDataMem_byte[4] = 0;
	printf("\tINFO: Waiting for HALT\r\n");
	prussdrv_pru_wait_event(PRU_EVTOUT_0);
	printf("\tINFO: PRU completed transfer\r\n");
	prussdrv_pru_clear_event(PRU_EVTOUT_0,PRU0_ARM_INTERRUPT);

	prussdrv_pru_disable(PRU_NUM);

	prussdrv_exit();
}
