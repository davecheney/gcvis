package main

type scvgtrace struct {
	inuse    int
	idle     int
	sys      int
	released int
	consumed int
}

type gctrace struct {
	NumGC       int
	Nproc       int
	t1          int
	t2          int
	t3          int
	t4          int
	Heap0       int // heap size before, in megabytes
	Heap1       int // heap size after, in megabytes
	Obj         int
	Goroutines  int
	NMalloc     int
	NFree       int
	NSpan       int
	NGoRoutines int
	NBGSweep    int
	NPauseSweep int
	NHandoff    int
	NHandoffCnt int
	NSteal      int
	NStealCnt   int
	NProcYield  int
	NOsYield    int
	NSleep      int
}
