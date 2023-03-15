package rangemap

import (
	"fmt"
	"strings"
)

// TODO: none of this should be in the translation package.

type Integral interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Range[E Integral] struct {
	Lo E
	Hi E
}

// Intersection returns the intersection of r2 with r, that is, a Range that
// contains only the values that are in both r and r2.
//
// Returns intersect, the intersection of r and r2, and valid, a boolean
// specifying whether the intersect is valid. If r and r2's intersection is not
// the empty set, valid will be true. If r and r2's intersection is the empty
// set (i.e. if they do not have any values in common), valid will be false and
// intersect should not be used.
func (r Range[E]) Intersection(r2 Range[E]) (intersect Range[E], valid bool) {
	if !r.Overlaps(r2) {
		return Range[E]{}, false
	}

	inter := r

	if r2.Lo > r.Lo {
		inter.Lo = r2.Lo
	}
	if r2.Hi < r.Hi {
		inter.Hi = r2.Hi
	}

	return inter, true
}

// Count returns the number of values included within the range.
func (r Range[E]) Count() int {
	return int(r.Hi - r.Lo + 1)
}

func (r Range[E]) String() string {
	return fmt.Sprintf("[%v, %v]", r.Lo, r.Hi)
}

func (r Range[E]) Contains(v E) bool {
	return v >= r.Lo && v <= r.Hi
}

// SubsetOf returns whether the set of values contained in r is a subset of the
// set of values contained in r2, that is, whether r1 is entirely within r2.
func (r Range[E]) SubsetOf(r2 Range[E]) bool {
	return r.Lo >= r2.Lo && r.Hi <= r2.Hi
}

// Overlaps returns whether any part of r overlaps with r2.
func (r Range[E]) Overlaps(r2 Range[E]) bool {
	// r overlaps r2 if either r.Lo or r.Hi are within r2.Lo and r2.Hi
	// or if r2.Lo or r2.Hi are within r.Lo and r.Hi.

	return r2.Contains(r.Lo) || r2.Contains(r.Hi) || r.Contains(r2.Lo) || r.Contains(r2.Hi)
}

// Returns the comparison of r to the other range. The returned value will be
// 0 if r == other, < 0 if r < other, and > 0 if r > other.
//
// Comparison is done as follows:
//
//   - r == r2 and Contains() returns 0 when r.Lo == r2.Lo and r.Hi == r2.Hi.
//
//   - r < r2 and Contains() returns < 0 when r.Lo < r2.Lo, or when
//     r.Lo == r2.Lo and r.Hi < r2.Hi.
//
//   - r > r2 and Contains() returns > 0 when r.Lo > r2.Lo, or when
//     r.Lo == r2.Lo and r.Hi > r2.Hi.
//
// Note that as a result of this comparison, if r2 is fully contained within r
// and r2.Lo > r.Lo, r is considered to be less than r2.
func (r Range[E]) Compare(r2 Range[E]) int {
	if r.Lo == r2.Lo {
		if r.Hi == r2.Hi {
			return 0
		} else if r.Hi < r2.Hi {
			return -1
		} else {
			return 1
		}
	} else if r.Lo < r2.Lo {
		return -1
	} else {
		return 1
	}
}

// RangeMap represents a piecewise function that maps ranges of its domain to
// ranges of its value range. The domain of the range map is always [0, n),
// where n is the value of Count().
//
// RangeMap has the property that the ranges added to it will always be mapped
// to in numerical order, as opposed to the order in which they were added. For
// example, adding (6, 10) to an empty RangeMap will result in a mapping of
// (0, 4) -> (6, 10), and if (2, 4) is then added, it will result in a mapping
// of (0, 2) -> (2, 4), (3, 7) -> (6, 10).
//
// The zero-value of RangeMap is an empty RangeMap ready for use.
type RangeMap[E Integral] struct {
	count   int
	ranges  []Range[E]
	domains []Range[E]
}

func (rm *RangeMap[E]) Copy() *RangeMap[E] {
	rmCopy := &RangeMap[E]{
		count:   rm.count,
		ranges:  make([]Range[E], len(rm.ranges)),
		domains: make([]Range[E], len(rm.domains)),
	}
	copy(rmCopy.ranges, rm.ranges)
	copy(rmCopy.domains, rm.domains)
	return rmCopy
}

// Intersection returns a new RangeMap that is the intersection of rm and rm2;
// that is, its mapped-to ranges contain only the values that are in both rm and
// rm2. If rm and rm2 do not have any values in common, the returned RangeMap
// will be empty.
func (rm *RangeMap[E]) Intersection(rm2 *RangeMap[E]) *RangeMap[E] {
	newRanges := make([]Range[E], 0)

	var lastIntersectedWith int
	// luckily, we can assume that the ranges are ordered for both maps.
	for i := 0; i < len(rm2.ranges); i++ {
		checkRange := rm2.ranges[i]
		var checkHasIntersected bool
		for j := lastIntersectedWith; j < len(rm.ranges); j++ {
			againstRange := rm.ranges[j]
			intersect, ok := checkRange.Intersection(againstRange)

			if ok {
				lastIntersectedWith = j
				checkHasIntersected = true
				newRanges = append(newRanges, intersect)
			} else if checkHasIntersected {
				// because the ranges are ordered, if our checkRange had
				// previously intersected but then didn't, we can stop checking
				// for that checkRange and continue with the next.
				//
				// However, the next checkRange *may* intersect with the one
				// that checkRange last intersected with, so we keep the
				// lastInteractedWith and track that.
				break
			}
		}
	}

	// now we have the list of new ranges, add them all to a new RangeMap
	intersectedMap := &RangeMap[E]{}

	for _, r := range newRanges {
		intersectedMap.Add(r.Lo, r.Hi)
	}

	return intersectedMap
}

func (rm *RangeMap[E]) String() string {
	var sb strings.Builder
	for i := range rm.ranges {
		sb.WriteRune('{')
		sb.WriteString(rm.ranges[i].String())
		sb.WriteRune('}')
		if i+1 < len(rm.ranges) {
			sb.WriteString(" U ")
		}
	}
	return fmt.Sprintf("RangeMap: {x | x ∈ [0, %d)} → {y | y ∈ %s}", rm.count, sb.String())
}

// Count returns the current number of values in the range map's domain. Count
// - 1 is the highest value allowed.
func (rm *RangeMap[E]) Count() int {
	return rm.count
}

// Call returns the value in the map that val maps to. val must be in the range
// [0, Count()). If E is not in the range, this function panics.
func (rm *RangeMap[E]) Call(val E) E {
	if val > E(rm.count) || val < 0 {
		panic(fmt.Sprintf("value outside domain [0, %d) of range map: %d", rm.count, val))
	}

	for i := range rm.domains {
		if rm.domains[i].Contains(val) {
			r := rm.ranges[i]
			domainToRangeShift := r.Lo - rm.domains[i].Lo
			return val + domainToRangeShift
		}
	}

	// should never happen
	panic(fmt.Sprintf("value has no mapping in range map: %d", val))
}

// Add ad.ds a range to the RangeMap. The range consists of the closed interval
// [start, end]. If the range is already in the RangeMap, this function has no
// effect.
func (rm *RangeMap[E]) Add(start, end E) {
	if rm.ranges == nil {
		rm.ranges = make([]Range[E], 0)
	}
	if rm.domains == nil {
		rm.domains = make([]Range[E], 0)
	}

	r := Range[E]{Lo: start, Hi: end}
	overlapping := rm.getRangesOverlapping(r)

	// make sure we are normalizing on exit
	defer rm.joinAdjacentRanges()

	// possible cases:

	// 1. r is entirely outside of all existing ranges
	//		- add r to the map regularly.
	if len(overlapping) == 0 {
		insertIdx := rm.findRangeInsertionPoint(r)

		// find r's domain
		var rDomain Range[E]
		if insertIdx > 0 {
			// if there is a previous range, r's domain starts exactly 1 after
			// its end. otherwise it starts at 0 (the default)
			rDomain.Lo = rm.domains[insertIdx-1].Hi + 1
		}
		rDomain.Hi = rDomain.Lo + (r.Hi - r.Lo)

		rm.insertMapping(insertIdx, rDomain, r)
		return
	}

	// 2. r is entirely contained within one existing range
	//		- in this case, do nothing. Range is already in the map.
	if len(overlapping) == 1 && r.SubsetOf(overlapping[0]) {
		return
	}

	// 3. start of r is inside an existing range, but end is outside all ranges
	//		- r may entirely cover one or more existing ranges besides the one
	//		it starts in.
	//		- extend the end of the existing range r starts in to be longer.
	//		For all ranges it ends up covering in the process of doing so,
	//		remove those ranges.
	if len(overlapping) >= 1 && overlapping[0].Contains(r.Lo) && !overlapping[len(overlapping)-1].Contains(r.Hi) {
		extendingRange := overlapping[0]

		// take the first range in the overlapping slice and extend it
		// to cover r.
		extendingIdx := rm.findRange(extendingRange)
		if extendingIdx == -1 {
			panic("overlapping range not found in map, should be impossible")
		}

		// get oldDomain before we remove the range
		oldDomain := rm.domains[extendingIdx]

		newRange := Range[E]{Lo: extendingRange.Lo, Hi: r.Hi}

		// remove all overlapping slices
		// TODO: efficiency. We could simply assume that we need to remove all
		// so once we find the index of the first one, we can just remove all
		// after that.
		for i := 0; i < len(overlapping); i++ {
			rm.removeRange(overlapping[i])
		}

		// now add the new one in
		insertIdx := rm.findRangeInsertionPoint(newRange)
		var domainLow E
		if insertIdx > 0 {
			// if there is a previous range, r's domain starts exactly 1 after
			// its end. otherwise it starts at 0 (the default)
			domainLow = rm.domains[insertIdx-1].Hi + 1
		}

		// start the new domain at zero so we can make future adjustments
		newDomain := Range[E]{Lo: 0, Hi: oldDomain.Hi - oldDomain.Lo}

		// find amount to extend domain by
		extensionAmt := r.Hi - extendingRange.Hi

		// adjust right side of domain to include r
		newDomain.Hi += extensionAmt

		// update domain to match new low
		domainDiff := newDomain.Lo - domainLow
		newDomain.Lo -= domainDiff
		newDomain.Hi -= domainDiff

		rm.insertMapping(insertIdx, newDomain, newRange)
		return
	}

	// 4. end of r is inside an existing range, but start is outside all ranges
	//		- r may entirely cover one or more existing ranges besides the one
	//		it ends in.
	//		- extend the start of the existing range r ends in to be longer.
	//		For all ranges it ends up covering in the process of doing so,
	//		remove those ranges.
	if len(overlapping) >= 1 && !overlapping[0].Contains(r.Lo) && overlapping[len(overlapping)-1].Contains(r.Hi) {
		extendingRange := overlapping[len(overlapping)-1]

		// take the first range in the overlapping slice and extend it
		// to cover r.
		extendingIdx := rm.findRange(extendingRange)
		if extendingIdx == -1 {
			panic("overlapping range not found in map, should be impossible")
		}

		// get oldDomain before we remove the range
		oldDomain := rm.domains[extendingIdx]

		newRange := Range[E]{Lo: r.Lo, Hi: extendingRange.Hi}

		// remove all overlapping slices
		// TODO: efficiency. We could simply assume that we need to remove all
		// so once we find the index of the first one, we can just remove all
		// after that.
		for i := 0; i < len(overlapping); i++ {
			rm.removeRange(overlapping[i])
		}

		insertIdx := rm.findRangeInsertionPoint(newRange)
		var domainLow E
		if insertIdx > 0 {
			// if there is a previous range, r's domain starts exactly 1 after
			// its end. otherwise it starts at 0 (the default)
			domainLow = rm.domains[insertIdx-1].Hi + 1
		}

		// start the new domain at zero so we can make future adjustments
		newDomain := Range[E]{Lo: 0, Hi: oldDomain.Hi - oldDomain.Lo}

		// find amount to extend domain by
		extensionAmt := extendingRange.Lo - r.Lo

		// adjust left side of domain to include r
		newDomain.Lo -= extensionAmt

		// update domain to match new low
		domainDiff := newDomain.Lo - domainLow
		newDomain.Lo -= domainDiff
		newDomain.Hi -= domainDiff

		rm.insertMapping(insertIdx, newDomain, newRange)
		return
	}

	// 5. start of r is inside an existing range, and end is inside another
	// existing range.
	//		- r may entirely cover one or more existing ranges besides the ones
	//		it starts and ends in.
	//		- join the two ranges r starts and ends in. For all ranges it ends
	//		up covering in the process of doing so, remove those ranges.
	if len(overlapping) >= 2 && overlapping[0].Contains(r.Lo) && overlapping[len(overlapping)-1].Contains(r.Hi) {
		// treat the start range as the extending range as we will start with that
		// one and add the end range to it to cover r.

		extendingRange := overlapping[0]
		endRange := overlapping[len(overlapping)-1]

		extendingIdx := rm.findRange(extendingRange)

		// get oldDomain before we remove the ranges
		oldDomain := rm.domains[extendingIdx]

		// join the two ranges r starts and ends in.
		newRange := Range[E]{Lo: extendingRange.Lo, Hi: endRange.Hi}

		// remove all overlapping slices
		// TODO: efficiency. We could simply assume that we need to remove all
		// so once we find the index of the first one, we can just remove all
		// after that.
		for i := 0; i < len(overlapping); i++ {
			rm.removeRange(overlapping[i])
		}

		// now add the new one in
		insertIdx := rm.findRangeInsertionPoint(newRange)
		var domainLow E
		if insertIdx > 0 {
			// if there is a previous range, r's domain starts exactly 1 after
			// its end. otherwise it starts at 0 (the default)
			domainLow = rm.domains[insertIdx-1].Hi + 1
		}

		// start the new domain at zero so we can make future adjustments
		// use left domain first (oldStartDomain)
		newDomain := Range[E]{Lo: 0, Hi: oldDomain.Hi - oldDomain.Lo}

		/*

			E1:   [------------]XXXXXXXXXXXXXX
			E2:                       [------]
			R:             [--------------]
			      e1L      rL  e1H    e2L rH e2H

		*/

		// find amount to extend domain by, so that both old domains are covered
		extensionAmt := endRange.Hi - extendingRange.Hi

		// adjust left side of domain to include r
		newDomain.Lo -= extensionAmt

		// update domain to match new low
		domainDiff := newDomain.Lo - domainLow
		newDomain.Lo -= domainDiff
		newDomain.Hi -= domainDiff

		rm.insertMapping(insertIdx, newDomain, newRange)
		return
	}

	// 6. start and end of r are both outside an existing range, but it entirely
	// covers one or more existing ranges.
	//		- remove all ranges r overlaps with, and add r normally.
	if len(overlapping) >= 1 && !overlapping[0].Contains(r.Lo) && !overlapping[len(overlapping)-1].Contains(r.Hi) {
		// remove all overlapping slices

		// TODO: efficiency. We could simply assume that we need to remove all
		// so once we find the index of the first one, we can just remove all
		// after that.
		for i := 0; i < len(overlapping); i++ {
			rm.removeRange(overlapping[i])
		}

		// insert r normally
		insertIdx := rm.findRangeInsertionPoint(r)

		// find r's domain
		var rDomain Range[E]
		if insertIdx > 0 {
			// if there is a previous range, r's domain starts exactly 1 after
			// its end. otherwise it starts at 0 (the default)
			rDomain.Lo = rm.domains[insertIdx-1].Hi + 1
		}
		rDomain.Hi = rDomain.Lo + (r.Hi - r.Lo)

		rm.insertMapping(insertIdx, rDomain, r)
		return
	}

	// we should never get here
	panic("unhandled case in RangeMap.Add")
}

// goes through and glues together any ranges that have ended up directly
// adjacent to each other.
func (rm *RangeMap[E]) joinAdjacentRanges() {
	newRanges := make([]Range[E], 0)
	newDomains := make([]Range[E], 0)

	for i := 0; i < len(rm.ranges); i++ {
		r := rm.ranges[i]
		d := rm.domains[i]

		// join until no more to join
		for i+1 < len(rm.ranges) && r.Hi+1 == rm.ranges[i+1].Lo {
			r.Hi = rm.ranges[i+1].Hi
			d.Hi = rm.domains[i+1].Hi
			i++ // skip the next one
		}

		newRanges = append(newRanges, r)
		newDomains = append(newDomains, d)
	}

	rm.ranges = newRanges
	rm.domains = newDomains
}

func (rm *RangeMap[E]) getRangesOverlapping(r Range[E]) []Range[E] {
	if rm.ranges == nil {
		return nil
	}

	var overlapping []Range[E]

	for _, check := range rm.ranges {
		if check.Overlaps(r) {
			overlapping = append(overlapping, check)
		}
	}

	return overlapping
}

func (rm *RangeMap[E]) findRangeInsertionPoint(r Range[E]) int {
	// find where to insert r in ordering
	insertIdx := -1
	for i, v := range rm.ranges {
		if v.Compare(r) > 0 {
			insertIdx = i
			break
		}
	}
	if insertIdx == -1 {
		// it is inserted at the end
		insertIdx = len(rm.ranges)
	}
	return insertIdx
}

// removeRange finds where the given range is in the mappings, and removes that
// domain and range, updating all domains that come after it to be lower.
func (rm *RangeMap[E]) removeRange(r Range[E]) {
	rIdx := rm.findRange(r)
	if rIdx == -1 {
		// it is already not in the map, nothing to do
		return
	}
	rm.removeMapping(rIdx)
}

// findRange returns the index of the given range in the mappings or -1 if it
// is not found.
func (rm *RangeMap[E]) findRange(r Range[E]) int {
	rIdx := -1
	for i, v := range rm.ranges {
		if v.Compare(r) == 0 {
			rIdx = i
			break
		}
	}
	return rIdx
}

// removeMapping removes the domain and range at the given index from the
// mapping and updates all domains that come after it to be lower.
//
// additionally, updates count.
func (rm *RangeMap[E]) removeMapping(idx int) {
	// need the domain for updating
	updateAmount := rm.domains[idx].Count()

	newRanges := make([]Range[E], len(rm.ranges)-1)
	newDomains := make([]Range[E], len(rm.domains)-1)
	copy(newRanges, rm.ranges[:idx])
	copy(newDomains, rm.domains[:idx])
	if idx < len(rm.ranges)-1 {
		// len(rm.ranges) and len(rm.domains) will always be the same
		copy(newRanges[idx:], rm.ranges[idx+1:])
		copy(newDomains[idx:], rm.domains[idx+1:])
	}
	rm.ranges = newRanges
	rm.domains = newDomains

	// need to update all domains that come after the one we just removed
	for i := idx; i < len(rm.domains); i++ {
		d := rm.domains[i]
		d.Lo -= E(updateAmount)
		d.Hi -= E(updateAmount)
		rm.domains[i] = d
	}

	rm.count -= updateAmount
}

// insert mapping adds the rDomain and rRange to the values at the given index,
// moves everyfin after that up to make room, and updates all domains that come
// after the inserted domain.
//
// additionally, updates count.
func (rm *RangeMap[E]) insertMapping(idx int, rDomain, rRange Range[E]) {
	// sanity check
	if rDomain.Count() != rRange.Count() {
		panic("domain and range to be inserted must be the same size")
	}

	newRanges := make([]Range[E], len(rm.ranges)+1)
	newDomains := make([]Range[E], len(rm.domains)+1)
	copy(newRanges, rm.ranges[:idx])
	copy(newDomains, rm.domains[:idx])
	newRanges[idx] = rRange
	newDomains[idx] = rDomain
	if idx < len(rm.ranges) {
		// len(rm.ranges) and len(rm.domains) will always be the same
		copy(newRanges[idx+1:], rm.ranges[idx:])
		copy(newDomains[idx+1:], rm.domains[idx:])
	}
	rm.ranges = newRanges
	rm.domains = newDomains

	// need to update all domains that come after the one we just inserted
	updateAmount := rDomain.Count()
	for i := idx + 1; i < len(rm.domains); i++ {
		d := rm.domains[i]
		d.Lo += E(updateAmount)
		d.Hi += E(updateAmount)
		rm.domains[i] = d
	}

	rm.count += updateAmount
}
