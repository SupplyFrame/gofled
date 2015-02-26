package main

/*
	transitions blend from one frame buffer to another
	transitions have an amount value which is the position through the transition from src to dst
	transition types:
		crossfade - lerps between src to dst using linear interpolation of each value
		dissolve - swaps out random values from src to dst in hard cuts, until all values are moved
		wipe - slides in the new values from one side, pushing the old values out to the other
*/