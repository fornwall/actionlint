package actionlint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestExprNewMapObjectType(t *testing.T) {
	o := NewMapObjectType(StringType{})
	if o.Props != nil {
		t.Fatalf("props should be nil but %v", o.Props)
	}
	if _, ok := o.Mapped.(StringType); !ok {
		t.Fatalf("mapped type is not string: %v", o.Mapped)
	}
	if o.IsStrict() {
		t.Fatalf("map object is not strict object but got %v", o)
	}
}

func TestExprObjectTypeSetStrict(t *testing.T) {
	o := NewEmptyObjectType()
	if o.IsStrict() || !o.IsLoose() {
		t.Fatal("should be loose")
	}
	o.Strict()
	if !o.IsStrict() || o.IsLoose() {
		t.Fatal("should be strict")
	}
	o.Loose()
	if o.IsStrict() || !o.IsLoose() {
		t.Fatal("should be loose")
	}
}

func TestExprAssignableSimple(t *testing.T) {
	testCases := []ExprType{
		AnyType{},
		NullType{},
		NumberType{},
		BoolType{},
		StringType{},
		NewObjectType(map[string]ExprType{"n": NumberType{}}),
		NewStrictObjectType(map[string]ExprType{"b": BoolType{}}),
		NewMapObjectType(NullType{}),
		&ArrayType{Elem: StringType{}},
	}

	for _, ty := range testCases {
		s := ty.String()
		t.Run(s, func(t *testing.T) {
			if !ty.Assignable(ty) {
				t.Fatalf("%s is not self-assignable", ty)
			}

			switch ty.(type) {
			case NullType:
			case AnyType:
			default:
				if (NullType{}).Assignable(ty) {
					t.Fatalf("%s is assignable to null", ty)
				}
			}

			if !(AnyType{}).Assignable(ty) {
				t.Fatalf("%s is not assignable to any", ty)
			}
		})
	}
}

func TestExprAssignableObject(t *testing.T) {
	testCases := []struct {
		from, to ExprType
		no       bool
	}{
		{
			from: NewMapObjectType(NumberType{}),
			to:   NewMapObjectType(StringType{}),
		},
		{
			from: NewEmptyObjectType(),
			to:   NewMapObjectType(StringType{}),
		},
		{
			from: NewMapObjectType(StringType{}),
			to:   NewEmptyObjectType(),
		},
		{
			from: NewStrictObjectType(map[string]ExprType{
				"a": NumberType{},
				"b": StringType{},
			}),
			to: NewMapObjectType(StringType{}),
		},
		{
			from: NewStrictObjectType(map[string]ExprType{"a": NullType{}}),
			to:   NewMapObjectType(StringType{}),
			no:   true,
		},
		{
			from: NewMapObjectType(NumberType{}),
			to: NewStrictObjectType(map[string]ExprType{
				"a": AnyType{},
				"b": StringType{},
			}),
		},
		{
			from: NewMapObjectType(NumberType{}),
			to: NewStrictObjectType(map[string]ExprType{
				"a": NullType{},
				"b": StringType{},
			}),
			no: true,
		},
		{
			from: NewStrictObjectType(map[string]ExprType{"a": NumberType{}}),
			to:   NewStrictObjectType(map[string]ExprType{"a": StringType{}}),
		},
		{
			from: NewStrictObjectType(map[string]ExprType{"a": StringType{}}),
			to:   NewStrictObjectType(map[string]ExprType{"b": StringType{}}),
			no:   true,
		},
		{
			from: NewStrictObjectType(map[string]ExprType{"a": NullType{}}),
			to:   NewStrictObjectType(map[string]ExprType{"a": StringType{}}),
			no:   true,
		},
	}

	for _, tc := range testCases {
		l, r := tc.to.String(), tc.from.String()
		t.Run(l+" := "+r, func(t *testing.T) {
			if tc.to.Assignable(tc.from) == tc.no {
				not := ""
				if tc.no {
					not = " not"
				}
				t.Fatalf("%s should%s be assignable to %s", r, not, l)
			}
		})
	}
}

func TestExprEqualTypes(t *testing.T) {
	testCases := []struct {
		what string
		ty   ExprType
		neq  ExprType
		eq   ExprType
	}{
		{
			what: "null",
			ty:   NullType{},
			neq:  StringType{},
		},
		{
			what: "number",
			ty:   NumberType{},
			neq:  StringType{},
		},
		{
			what: "bool",
			ty:   BoolType{},
			neq:  StringType{},
		},
		{
			what: "string",
			ty:   StringType{},
			neq:  BoolType{},
		},
		{
			what: "object",
			ty:   NewEmptyObjectType(),
			neq:  &ArrayType{Elem: AnyType{}},
		},
		{
			what: "strict props object",
			ty:   NewEmptyStrictObjectType(),
			neq:  &ArrayType{Elem: AnyType{}},
		},
		{
			what: "nested object",
			ty: NewObjectType(map[string]ExprType{
				"foo": NewObjectType(map[string]ExprType{
					"bar": StringType{},
				}),
			}),
			neq: &ArrayType{Elem: AnyType{}},
		},
		{
			what: "nested strict props object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": NewStrictObjectType(map[string]ExprType{
					"bar": StringType{},
				}),
			}),
			neq: &ArrayType{Elem: AnyType{}},
		},
		{
			what: "nested object prop name",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			neq: NewStrictObjectType(map[string]ExprType{
				"bar": StringType{},
			}),
		},
		{
			what: "nested object prop type",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			neq: NewStrictObjectType(map[string]ExprType{
				"foo": BoolType{},
			}),
		},
		{
			what: "strict props object and loose object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": NullType{},
			}),
			eq: NewObjectType(map[string]ExprType{
				"foo": NullType{},
			}),
		},
		{
			what: "loose object and strict props object",
			ty: NewObjectType(map[string]ExprType{
				"foo": NullType{},
			}),
			eq: NewObjectType(map[string]ExprType{
				"foo": NullType{},
			}),
		},
		{
			what: "map objects",
			ty:   NewMapObjectType(NullType{}),
			eq:   NewMapObjectType(NullType{}),
			neq:  NewMapObjectType(NumberType{}),
		},
		{
			what: "map object equals loose object",
			ty:   NewMapObjectType(StringType{}),
			eq:   NewEmptyObjectType(),
		},
		{
			what: "loose object equals map object",
			ty:   NewEmptyObjectType(),
			eq:   NewMapObjectType(StringType{}),
		},
		{
			what: "map object equals strict object",
			ty:   NewMapObjectType(StringType{}),
			eq: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			neq: NewStrictObjectType(map[string]ExprType{
				"foo": NullType{},
			}),
		},
		{
			what: "map object equals strict object including any prop",
			ty:   NewMapObjectType(StringType{}),
			eq: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
				"bar": AnyType{},
			}),
			neq: NewStrictObjectType(map[string]ExprType{
				"foo": NullType{},
				"bar": AnyType{},
			}),
		},
		{
			what: "strict object equals map object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			eq:  NewMapObjectType(StringType{}),
			neq: NewMapObjectType(NullType{}),
		},
		{
			what: "array",
			ty:   &ArrayType{Elem: StringType{}},
			neq:  NewEmptyObjectType(),
		},
		{
			what: "array element type",
			ty:   &ArrayType{Elem: StringType{}},
			neq:  &ArrayType{Elem: BoolType{}},
		},
		{
			what: "nested array",
			ty: &ArrayType{
				Elem: &ArrayType{Elem: StringType{}},
			},
			neq: &ArrayType{
				Elem: &ArrayType{Elem: BoolType{}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			var l, r ExprType

			l, r = tc.ty, tc.ty
			if !EqualTypes(l, r) {
				t.Errorf("%s should equal to %s", l.String(), r.String())
			}
			if tc.neq != nil {
				l, r = tc.ty, tc.neq
				if EqualTypes(l, r) {
					t.Errorf("%s should not equal to %s", l.String(), r.String())
				}
			}
			if tc.eq != nil {
				l, r = tc.ty, tc.eq
				if !EqualTypes(l, r) {
					t.Errorf("%s should equal to %s", l.String(), r.String())
				}
			}
			l, r = tc.ty, AnyType{}
			if !EqualTypes(l, r) {
				t.Errorf("%s should equal to %s", l.String(), r.String())
			}
			l, r = AnyType{}, tc.ty
			if !EqualTypes(l, r) {
				t.Errorf("%s should equal to %s", l.String(), r.String())
			}
		})
	}
}

func TestExprTypeStringize(t *testing.T) {
	testCases := []struct {
		what string
		ty   ExprType
		want string
	}{
		{
			what: "any",
			ty:   AnyType{},
			want: "any",
		},
		{
			what: "null",
			ty:   NullType{},
			want: "null",
		},
		{
			what: "number",
			ty:   NumberType{},
			want: "number",
		},
		{
			what: "bool",
			ty:   BoolType{},
			want: "bool",
		},
		{
			what: "string",
			ty:   StringType{},
			want: "string",
		},
		{
			what: "empty object",
			ty:   NewEmptyObjectType(),
			want: "object",
		},
		{
			what: "empty strict props object",
			ty:   NewEmptyStrictObjectType(),
			want: "{}",
		},
		{
			what: "strict object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			want: "{foo: string}",
		},
		{
			what: "non-strict object",
			ty: NewObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			want: "object",
		},
		{
			what: "strict props object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			want: "{foo: string}",
		},
		{
			what: "array",
			ty:   &ArrayType{Elem: AnyType{}},
			want: "array<any>",
		},
		{
			what: "nested array",
			ty:   &ArrayType{Elem: &ArrayType{BoolType{}, true}},
			want: "array<array<bool>>",
		},
		{
			what: "object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": &ArrayType{
					Elem: NewStrictObjectType(map[string]ExprType{
						"bar": &ArrayType{
							Elem: StringType{},
						},
					}),
				},
			}),
			want: "{foo: array<{bar: array<string>}>}",
		},
		{
			what: "map object",
			ty:   NewMapObjectType(NumberType{}),
			want: "{string => number}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			have := tc.ty.String()
			if have != tc.want {
				t.Fatalf("wanted %q but got %q", tc.want, have)
			}
		})
	}
}

func TestExprTypeMergeSimple(t *testing.T) {
	testCases := []ExprType{
		AnyType{},
		NullType{},
		NumberType{},
		BoolType{},
		StringType{},
		NewEmptyObjectType(),
		NewEmptyStrictObjectType(),
		NewMapObjectType(NullType{}),
		&ArrayType{Elem: StringType{}},
	}

	opt := cmpopts.EquateEmpty()

	for _, ty := range testCases {
		t.Run("any/"+ty.String(), func(t *testing.T) {
			have := ty.Merge(AnyType{})
			if _, ok := have.(AnyType); !ok {
				t.Errorf("any type merged with %s was %s while expecting any", ty.String(), have.String())
			}

			have = (AnyType{}).Merge(ty)
			if _, ok := have.(AnyType); !ok {
				t.Errorf("%s merged with any type was %s while expecting any", ty.String(), have.String())
			}
		})
	}

	for _, ty := range testCases {
		t.Run("incompatible/"+ty.String(), func(t *testing.T) {
			var in ExprType
			in = NullType{}
			if ty == (NullType{}) {
				in = StringType{} // null is compatible with null so use string instead
			}

			have := ty.Merge(in)
			if _, ok := have.(AnyType); !ok {
				t.Errorf("incompatible %s type merged with %s was %s while expecting any", in.String(), ty.String(), have.String())
			}
		})
	}

	for _, ty := range testCases {
		t.Run("self/"+ty.String(), func(t *testing.T) {
			have := ty.Merge(ty)
			if !cmp.Equal(ty, have, opt) {
				s := ty.String()
				t.Errorf("%s merged with %s was %s while expecting %s", s, s, have.String(), s)
			}
		})
	}
}

func TestExprTypeMergeComplicated(t *testing.T) {
	testCases := []struct {
		what string
		ty   ExprType
		with ExprType
		want ExprType
	}{
		{
			what: "number merges with string",
			ty:   NumberType{},
			with: StringType{},
			want: StringType{},
		},
		{
			what: "string is merged by number",
			ty:   StringType{},
			with: NumberType{},
			want: StringType{},
		},
		{
			what: "bool merges with string",
			ty:   BoolType{},
			with: StringType{},
			want: StringType{},
		},
		{
			what: "string is merged by bool",
			ty:   StringType{},
			with: BoolType{},
			want: StringType{},
		},
		{
			what: "object props",
			ty: NewObjectType(map[string]ExprType{
				"foo": NumberType{},
			}),
			with: NewObjectType(map[string]ExprType{
				"bar": StringType{},
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": NumberType{},
				"bar": StringType{},
			}),
		},
		{
			what: "loose object with strict object",
			ty: NewObjectType(map[string]ExprType{
				"foo": NumberType{},
			}),
			with: NewStrictObjectType(map[string]ExprType{
				"bar": StringType{},
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": NumberType{},
				"bar": StringType{},
			}),
		},
		{
			what: "strict object with strict object",
			ty: NewStrictObjectType(map[string]ExprType{
				"foo": NumberType{},
			}),
			with: NewStrictObjectType(map[string]ExprType{
				"bar": StringType{},
			}),
			want: NewStrictObjectType(map[string]ExprType{
				"foo": NumberType{},
				"bar": StringType{},
			}),
		},
		{
			what: "compatible prop",
			ty: NewObjectType(map[string]ExprType{
				"foo": NumberType{},
			}),
			with: NewObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
		},
		{
			what: "any prop with prop",
			ty: NewObjectType(map[string]ExprType{
				"foo": AnyType{},
			}),
			with: NewObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": AnyType{},
			}),
		},
		{
			what: "prop with any prop",
			ty: NewObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			with: NewObjectType(map[string]ExprType{
				"foo": AnyType{},
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": AnyType{},
			}),
		},
		{
			what: "incompatible prop",
			ty: NewObjectType(map[string]ExprType{
				"foo": NullType{},
			}),
			with: NewObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": AnyType{},
			}),
		},
		{
			what: "compatible array element",
			ty: &ArrayType{
				Elem: NumberType{},
			},
			with: &ArrayType{
				Elem: StringType{},
			},
			want: &ArrayType{
				Elem: StringType{},
			},
		},
		{
			what: "incompatible array element",
			ty: &ArrayType{
				Elem: NullType{},
			},
			with: &ArrayType{
				Elem: StringType{},
			},
			want: &ArrayType{
				Elem: AnyType{},
			},
		},
		{
			what: "any array element with element",
			ty: &ArrayType{
				Elem: AnyType{},
			},
			with: &ArrayType{
				Elem: StringType{},
			},
			want: &ArrayType{
				Elem: AnyType{},
			},
		},
		{
			what: "array element with any element",
			ty: &ArrayType{
				Elem: StringType{},
			},
			with: &ArrayType{
				Elem: AnyType{},
			},
			want: &ArrayType{
				Elem: AnyType{},
			},
		},
		{
			what: "array with array deref",
			ty:   &ArrayType{StringType{}, false},
			with: &ArrayType{StringType{}, true},
			want: &ArrayType{StringType{}, false},
		},
		{
			what: "array deref with array",
			ty:   &ArrayType{StringType{}, true},
			with: &ArrayType{StringType{}, false},
			want: &ArrayType{StringType{}, false},
		},
		{
			what: "array deref with array deref",
			ty:   &ArrayType{StringType{}, true},
			with: &ArrayType{StringType{}, true},
			want: &ArrayType{StringType{}, false},
		},
		{
			what: "object no prop at left hand side",
			ty:   NewEmptyObjectType(),
			with: NewObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
		},
		{
			what: "object no prop at right hand side",
			ty: NewObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
			with: NewEmptyObjectType(),
			want: NewObjectType(map[string]ExprType{
				"foo": StringType{},
			}),
		},
		{
			what: "any elem array at left hand side",
			ty:   &ArrayType{AnyType{}, false},
			with: &ArrayType{StringType{}, false},
			want: &ArrayType{AnyType{}, false},
		},
		{
			what: "any elem array at right hand side",
			ty:   &ArrayType{StringType{}, false},
			with: &ArrayType{AnyType{}, false},
			want: &ArrayType{AnyType{}, false},
		},
		{
			what: "nested array",
			ty: &ArrayType{
				Elem: &ArrayType{
					Elem: NumberType{},
				},
			},
			with: &ArrayType{
				Elem: &ArrayType{
					Elem: StringType{},
				},
			},
			want: &ArrayType{
				Elem: &ArrayType{
					Elem: StringType{},
				},
			},
		},
		{
			what: "nested objects",
			ty: NewObjectType(map[string]ExprType{
				"foo": NewObjectType(map[string]ExprType{
					"foo":  NumberType{},
					"piyo": NumberType{},
				}),
				"aaa": NumberType{},
				"ccc": NumberType{},
			}),
			with: NewObjectType(map[string]ExprType{
				"foo": NewObjectType(map[string]ExprType{
					"bar":  StringType{},
					"piyo": StringType{},
				}),
				"bbb": StringType{},
				"ccc": StringType{},
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": NewObjectType(map[string]ExprType{
					"foo":  NumberType{},
					"bar":  StringType{},
					"piyo": StringType{},
				}),
				"aaa": NumberType{},
				"bbb": StringType{},
				"ccc": StringType{},
			}),
		},
		{
			what: "map object with compatible map object",
			ty:   NewMapObjectType(NumberType{}),
			with: NewMapObjectType(StringType{}),
			want: NewMapObjectType(StringType{}),
		},
		{
			what: "map object with incompatible map object",
			ty:   NewMapObjectType(NumberType{}),
			with: NewMapObjectType(NullType{}),
			want: NewEmptyObjectType(),
		},
		{
			what: "map object with compatible object",
			ty:   NewMapObjectType(NumberType{}),
			with: NewObjectType(map[string]ExprType{
				"foo": NumberType{},
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": NumberType{},
			}),
		},
		{
			what: "map object with incompatible object",
			ty:   NewMapObjectType(NumberType{}),
			with: NewObjectType(map[string]ExprType{
				"foo": BoolType{},
			}),
			want: NewObjectType(map[string]ExprType{
				"foo": BoolType{},
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.what, func(t *testing.T) {
			opt := cmpopts.EquateEmpty()
			ty := tc.with.Merge(tc.ty)
			if !cmp.Equal(ty, tc.want, opt) {
				t.Fatalf(
					"%s was merged with %s as %s while expecting %s\ndiff:\n%s",
					tc.ty.String(),
					tc.with.String(),
					ty.String(),
					tc.want.String(),
					cmp.Diff(tc.want, ty, opt),
				)
			}
		})
	}
}

func TestExprTypeMergeCreateNewInstance(t *testing.T) {
	{
		ty := &ArrayType{
			Elem: NumberType{},
		}
		ty2 := ty.Merge(&ArrayType{
			Elem: StringType{},
		})
		if ty == ty2 {
			t.Fatalf("did not make a new instance (%v => %v)", ty, ty2)
		}
		if _, ok := ty.Elem.(NumberType); !ok {
			t.Fatalf("original element type was modified: %v", ty)
		}
	}

	{
		ty := NewObjectType(map[string]ExprType{
			"foo": NumberType{},
		})
		ty2 := ty.Merge(
			NewObjectType(map[string]ExprType{
				"foo": StringType{},
				"bar": BoolType{},
			}),
		)
		if ty == ty2 {
			t.Fatalf("did not make a new instance (%v => %v)", ty, ty2)
		}
		if len(ty.Props) != 1 {
			t.Fatalf("new prop was added: %v", ty)
		}
		if _, ok := ty.Props["foo"].(NumberType); !ok {
			t.Fatalf("prop type was modified: %v", ty)
		}
	}
}
