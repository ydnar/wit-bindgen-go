package cm

import (
	"encoding/json"
	"errors"
)

// ErrInvalidTuple is returned when a Tuple fails to unmarshal from JSON.
var ErrInvalidTuple = errors.New("invalid tuple")

// Tuple represents a [Component Model tuple] with 2 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple[T0, T1 any] struct {
	_  HostLayout
	F0 T0
	F1 T1
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple[T0, T1]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple[T0, T1]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 2 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	return nil
}

// Tuple3 represents a [Component Model tuple] with 3 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple3[T0, T1, T2 any] struct {
	_  HostLayout
	F0 T0
	F1 T1
	F2 T2
}

func (t Tuple3[T0, T1, T2]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple3[T0, T1, T2]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 3 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	return nil
}

// Tuple4 represents a [Component Model tuple] with 4 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple4[T0, T1, T2, T3 any] struct {
	_  HostLayout
	F0 T0
	F1 T1
	F2 T2
	F3 T3
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple4[T0, T1, T2, T3]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple4[T0, T1, T2, T3]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 4 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	return nil
}

// Tuple5 represents a [Component Model tuple] with 5 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple5[T0, T1, T2, T3, T4 any] struct {
	_  HostLayout
	F0 T0
	F1 T1
	F2 T2
	F3 T3
	F4 T4
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple5[T0, T1, T2, T3, T4]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3, t.F4}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple5[T0, T1, T2, T3, T4]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 5 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &t.F4); err != nil {
		return err
	}
	return nil
}

// Tuple6 represents a [Component Model tuple] with 6 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple6[T0, T1, T2, T3, T4, T5 any] struct {
	_  HostLayout
	F0 T0
	F1 T1
	F2 T2
	F3 T3
	F4 T4
	F5 T5
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple6[T0, T1, T2, T3, T4, T5]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3, t.F4, t.F5}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple6[T0, T1, T2, T3, T4, T5]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 6 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &t.F4); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[5], &t.F5); err != nil {
		return err
	}
	return nil
}

// Tuple7 represents a [Component Model tuple] with 7 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple7[T0, T1, T2, T3, T4, T5, T6 any] struct {
	_  HostLayout
	F0 T0
	F1 T1
	F2 T2
	F3 T3
	F4 T4
	F5 T5
	F6 T6
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple7[T0, T1, T2, T3, T4, T5, T6]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3, t.F4, t.F5, t.F6}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple7[T0, T1, T2, T3, T4, T5, T6]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 7 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &t.F4); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[5], &t.F5); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[6], &t.F6); err != nil {
		return err
	}
	return nil
}

// Tuple8 represents a [Component Model tuple] with 8 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple8[T0, T1, T2, T3, T4, T5, T6, T7 any] struct {
	_  HostLayout
	F0 T0
	F1 T1
	F2 T2
	F3 T3
	F4 T4
	F5 T5
	F6 T6
	F7 T7
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3, t.F4, t.F5, t.F6, t.F7}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 8 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &t.F4); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[5], &t.F5); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[6], &t.F6); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[7], &t.F7); err != nil {
		return err
	}
	return nil
}

// Tuple9 represents a [Component Model tuple] with 9 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8 any] struct {
	_  HostLayout
	F0 T0
	F1 T1
	F2 T2
	F3 T3
	F4 T4
	F5 T5
	F6 T6
	F7 T7
	F8 T8
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3, t.F4, t.F5, t.F6, t.F7, t.F8}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 9 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &t.F4); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[5], &t.F5); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[6], &t.F6); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[7], &t.F7); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[8], &t.F8); err != nil {
		return err
	}
	return nil
}

// Tuple10 represents a [Component Model tuple] with 10 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9 any] struct {
	_  HostLayout
	F0 T0
	F1 T1
	F2 T2
	F3 T3
	F4 T4
	F5 T5
	F6 T6
	F7 T7
	F8 T8
	F9 T9
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3, t.F4, t.F5, t.F6, t.F7, t.F8, t.F9}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 10 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &t.F4); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[5], &t.F5); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[6], &t.F6); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[7], &t.F7); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[8], &t.F8); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[9], &t.F9); err != nil {
		return err
	}
	return nil
}

// Tuple11 represents a [Component Model tuple] with 11 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any] struct {
	_   HostLayout
	F0  T0
	F1  T1
	F2  T2
	F3  T3
	F4  T4
	F5  T5
	F6  T6
	F7  T7
	F8  T8
	F9  T9
	F10 T10
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3, t.F4, t.F5, t.F6, t.F7, t.F8, t.F9, t.F10}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 11 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &t.F4); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[5], &t.F5); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[6], &t.F6); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[7], &t.F7); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[8], &t.F8); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[9], &t.F9); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[10], &t.F10); err != nil {
		return err
	}
	return nil
}

// Tuple12 represents a [Component Model tuple] with 12 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any] struct {
	_   HostLayout
	F0  T0
	F1  T1
	F2  T2
	F3  T3
	F4  T4
	F5  T5
	F6  T6
	F7  T7
	F8  T8
	F9  T9
	F10 T10
	F11 T11
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3, t.F4, t.F5, t.F6, t.F7, t.F8, t.F9, t.F10, t.F11}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 12 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &t.F4); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[5], &t.F5); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[6], &t.F6); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[7], &t.F7); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[8], &t.F8); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[9], &t.F9); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[10], &t.F10); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[11], &t.F11); err != nil {
		return err
	}
	return nil
}

// Tuple13 represents a [Component Model tuple] with 13 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any] struct {
	_   HostLayout
	F0  T0
	F1  T1
	F2  T2
	F3  T3
	F4  T4
	F5  T5
	F6  T6
	F7  T7
	F8  T8
	F9  T9
	F10 T10
	F11 T11
	F12 T12
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3, t.F4, t.F5, t.F6, t.F7, t.F8, t.F9, t.F10, t.F11, t.F12}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 13 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &t.F4); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[5], &t.F5); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[6], &t.F6); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[7], &t.F7); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[8], &t.F8); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[9], &t.F9); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[10], &t.F10); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[11], &t.F11); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[12], &t.F12); err != nil {
		return err
	}
	return nil
}

// Tuple14 represents a [Component Model tuple] with 14 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any] struct {
	_   HostLayout
	F0  T0
	F1  T1
	F2  T2
	F3  T3
	F4  T4
	F5  T5
	F6  T6
	F7  T7
	F8  T8
	F9  T9
	F10 T10
	F11 T11
	F12 T12
	F13 T13
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3, t.F4, t.F5, t.F6, t.F7, t.F8, t.F9, t.F10, t.F11, t.F12, t.F13}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 14 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &t.F4); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[5], &t.F5); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[6], &t.F6); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[7], &t.F7); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[8], &t.F8); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[9], &t.F9); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[10], &t.F10); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[11], &t.F11); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[12], &t.F12); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[13], &t.F13); err != nil {
		return err
	}
	return nil
}

// Tuple15 represents a [Component Model tuple] with 15 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any] struct {
	_   HostLayout
	F0  T0
	F1  T1
	F2  T2
	F3  T3
	F4  T4
	F5  T5
	F6  T6
	F7  T7
	F8  T8
	F9  T9
	F10 T10
	F11 T11
	F12 T12
	F13 T13
	F14 T14
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3, t.F4, t.F5, t.F6, t.F7, t.F8, t.F9, t.F10, t.F11, t.F12, t.F13, t.F14}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 15 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &t.F4); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[5], &t.F5); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[6], &t.F6); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[7], &t.F7); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[8], &t.F8); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[9], &t.F9); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[10], &t.F10); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[11], &t.F11); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[12], &t.F12); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[13], &t.F13); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[14], &t.F14); err != nil {
		return err
	}
	return nil
}

// Tuple16 represents a [Component Model tuple] with 16 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any] struct {
	_   HostLayout
	F0  T0
	F1  T1
	F2  T2
	F3  T3
	F4  T4
	F5  T5
	F6  T6
	F7  T7
	F8  T8
	F9  T9
	F10 T10
	F11 T11
	F12 T12
	F13 T13
	F14 T14
	F15 T15
}

// MarshalJSON marshals the Tuple into JSON.
func (t Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T16]) MarshalJSON() ([]byte, error) {
	l := []any{t.F0, t.F1, t.F2, t.F3, t.F4, t.F5, t.F6, t.F7, t.F8, t.F9, t.F10, t.F11, t.F12, t.F13, t.F14, t.F15}
	return json.Marshal(l)
}

// UnmarshalJSON unmarshals the Tuple from JSON.
func (t *Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) UnmarshalJSON(buf []byte) error {
	tmp := []json.RawMessage{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 16 {
		return ErrInvalidTuple
	}
	if err := json.Unmarshal(tmp[0], &t.F0); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &t.F1); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &t.F2); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &t.F3); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &t.F4); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[5], &t.F5); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[6], &t.F6); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[7], &t.F7); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[8], &t.F8); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[9], &t.F9); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[10], &t.F10); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[11], &t.F11); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[12], &t.F12); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[13], &t.F13); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[14], &t.F14); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[15], &t.F15); err != nil {
		return err
	}
	return nil
}

// MaxTuple specifies the maximum number of fields in a Tuple* type, currently [Tuple16].
// See https://github.com/WebAssembly/component-model/issues/373 for more information.
const MaxTuple = 16
