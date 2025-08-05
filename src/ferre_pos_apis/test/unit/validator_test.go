package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ferre_pos_apis/pkg/validator"
)

func TestValidatorNew(t *testing.T) {
	v := validator.New()
	assert.NotNil(t, v)
}

func TestValidateStruct(t *testing.T) {
	v := validator.New()

	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
		Age   int    `validate:"min=0,max=120"`
	}

	tests := []struct {
		name        string
		input       TestStruct
		expectError bool
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   30,
			},
			expectError: false,
		},
		{
			name: "missing required field",
			input: TestStruct{
				Name:  "",
				Email: "john@example.com",
				Age:   30,
			},
			expectError: true,
		},
		{
			name: "invalid email",
			input: TestStruct{
				Name:  "John Doe",
				Email: "invalid-email",
				Age:   30,
			},
			expectError: true,
		},
		{
			name: "age out of range",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   150,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateField(t *testing.T) {
	v := validator.New()

	tests := []struct {
		name        string
		field       interface{}
		tag         string
		expectError bool
	}{
		{
			name:        "valid required field",
			field:       "test",
			tag:         "required",
			expectError: false,
		},
		{
			name:        "empty required field",
			field:       "",
			tag:         "required",
			expectError: true,
		},
		{
			name:        "valid email",
			field:       "test@example.com",
			tag:         "email",
			expectError: false,
		},
		{
			name:        "invalid email",
			field:       "invalid-email",
			tag:         "email",
			expectError: true,
		},
		{
			name:        "valid min value",
			field:       10,
			tag:         "min=5",
			expectError: false,
		},
		{
			name:        "invalid min value",
			field:       3,
			tag:         "min=5",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateField(tt.field, tt.tag)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetValidationErrors(t *testing.T) {
	v := validator.New()

	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
	}

	input := TestStruct{
		Name:  "",
		Email: "invalid-email",
	}

	err := v.ValidateStruct(input)
	assert.Error(t, err)

	validationErrors := v.GetValidationErrors(err)
	assert.NotEmpty(t, validationErrors)

	// Verificar que se devuelven errores estructurados
	for _, validationError := range validationErrors {
		assert.NotEmpty(t, validationError.Field)
		assert.NotEmpty(t, validationError.Tag)
		assert.NotEmpty(t, validationError.Message)
	}
}

func TestCustomValidations(t *testing.T) {
	v := validator.New()

	// Test de validación personalizada (si existe)
	type ProductStruct struct {
		Code string `validate:"required"`
		Name string `validate:"required"`
	}

	product := ProductStruct{
		Code: "PROD001",
		Name: "Test Product",
	}

	err := v.ValidateStruct(product)
	assert.NoError(t, err)
}

func BenchmarkValidateStruct(b *testing.B) {
	v := validator.New()

	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
		Age   int    `validate:"min=0,max=120"`
	}

	input := TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.ValidateStruct(input)
	}
}

func BenchmarkValidateField(b *testing.B) {
	v := validator.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.ValidateField("test@example.com", "email")
	}
}

