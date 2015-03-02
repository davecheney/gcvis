package main

type scvgtrace struct {
	inuse    int64
	idle     int64
	sys      int64
	released int64
	consumed int64
}

type gctrace struct {
	NumGC       int64
	Nproc       int64
	t1          int64
	t2          int64
	t3          int64
	t4          int64
	Heap0       int64 // heap size before, in megabytes
	Heap1       int64 // heap size after, in megabytes
	Obj         int64
	NMalloc     int64
	NFree       int64
	NSpan       int64
	NGoRoutines int64
	NBGSweep    int64
	NPauseSweep int64
	NHandoff    int64
	NHandoffCnt int64
	NSteal      int64
	NStealCnt   int64
	NProcYield  int64
	NOsYield    int64
	NSleep      int64
}
