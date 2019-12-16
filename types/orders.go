package types

type Orders []*Order

func (os Orders) Delete(key int) []*Order {
	if key < len(os)-1 {
		copy(os[key:], os[key+1:])
	}
	os[len(os)-1] = nil

	return os[:len(os)-1]
}

func (os Orders) Insert(key int, order *Order) []*Order {
	os = append(os, &Order{})
	copy(os[key+1:], os[key:])
	os[key] = order

	return os
}
