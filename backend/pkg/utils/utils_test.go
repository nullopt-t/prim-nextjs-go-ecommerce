package utils_test

import (
	"testing"

	"github.com/nullopt-t/prim-backend/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidUUID(t *testing.T) {
	assert.False(t, utils.IsValidUUID(""))
	assert.False(t, utils.IsValidUUID("not-a-uuid"))
	assert.False(t, utils.IsValidUUID("00000000-0000-0000-0000-000000000000"))
	assert.True(t, utils.IsValidUUID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"))
	assert.True(t, utils.IsValidAnyUUID("00000000-0000-0000-0000-000000000000"))
}

func TestStringToIntAndParseInt(t *testing.T) {
	assert.Equal(t, 42, utils.StringToInt("42", 0))
	assert.Equal(t, 10, utils.StringToInt("invalid", 10))

	val, err := utils.ParseInt("123")
	require.NoError(t, err)
	assert.Equal(t, 123, val)

	_, err = utils.ParseInt("invalid")
	require.Error(t, err)
}

func TestStringToFloatAndParseFloat(t *testing.T) {
	assert.Equal(t, 3.14, utils.StringToFloat("3.14", 0.0))
	assert.Equal(t, 1.0, utils.StringToFloat("invalid", 1.0))
	assert.Equal(t, 1.0, utils.StringToFloat("NaN", 1.0))
	assert.Equal(t, 1.0, utils.StringToFloat("+Inf", 1.0))
	assert.Equal(t, 1.0, utils.StringToFloat("-Inf", 1.0))

	val, err := utils.ParseFloat("2.718")
	require.NoError(t, err)
	assert.Equal(t, 2.718, val)

	_, err = utils.ParseFloat("NaN")
	require.Error(t, err)

	_, err = utils.ParseFloat("+Inf")
	require.Error(t, err)
}

func TestStringToBoolAndParseBool(t *testing.T) {
	assert.True(t, utils.StringToBool("true", false))
	assert.False(t, utils.StringToBool("false", true))
	assert.True(t, utils.StringToBool("invalid", true))

	val, err := utils.ParseBool("true")
	require.NoError(t, err)
	assert.True(t, val)

	_, err = utils.ParseBool("invalid")
	require.Error(t, err)
}

func TestIntToString(t *testing.T) {
	assert.Equal(t, "42", utils.IntToString(42))
	assert.Equal(t, "-10", utils.IntToString(-10))
}

func TestEnvHelpers(t *testing.T) {
	t.Setenv("TEST_INT", "100")
	t.Setenv("TEST_INT_EMPTY", "   ")
	t.Setenv("TEST_INT_INVALID", "abc")

	assert.Equal(t, 100, utils.GetEnvAsInt("TEST_INT", 0))
	assert.Equal(t, 5, utils.GetEnvAsInt("TEST_INT_EMPTY", 5))
	assert.Equal(t, 10, utils.GetEnvAsInt("TEST_INT_INVALID", 10))
	assert.Equal(t, 20, utils.GetEnvAsInt("NON_EXISTENT", 20))

	t.Setenv("TEST_BOOL_YES", "yes")
	t.Setenv("TEST_BOOL_OFF", "off")
	t.Setenv("TEST_BOOL_EMPTY", "")
	t.Setenv("TEST_BOOL_INVALID", "maybe")

	assert.True(t, utils.GetEnvAsBool("TEST_BOOL_YES", false))
	assert.False(t, utils.GetEnvAsBool("TEST_BOOL_OFF", true))
	assert.True(t, utils.GetEnvAsBool("TEST_BOOL_EMPTY", true))
	assert.False(t, utils.GetEnvAsBool("TEST_BOOL_INVALID", false))

	t.Setenv("TEST_STR", "hello")
	t.Setenv("TEST_STR_EMPTY", "   ")

	assert.Equal(t, "hello", utils.GetEnvOrDefault("TEST_STR", "default"))
	assert.Equal(t, "default", utils.GetEnvOrDefault("TEST_STR_EMPTY", "default"))
	assert.Equal(t, "default", utils.GetEnvOrDefault("NON_EXISTENT_STR", "default"))
	assert.Equal(t, "default", utils.GetEnvNonEmptyOrDefault("TEST_STR_EMPTY", "default"))
}

func TestIsValidPhone(t *testing.T) {
	// International with +
	normalized, err := utils.IsValidPhone("+12025550123")
	require.NoError(t, err)
	assert.Equal(t, "+12025550123", normalized)

	// International with 00 prefix
	normalized, err = utils.IsValidPhone("0012025550123")
	require.NoError(t, err)
	assert.Equal(t, "+12025550123", normalized)

	// Country code without +
	normalized, err = utils.IsValidPhone("12025550123")
	require.NoError(t, err)
	assert.Equal(t, "+12025550123", normalized)

	// Local number with region
	normalized, err = utils.IsValidPhone("2025550123", "US")
	require.NoError(t, err)
	assert.Equal(t, "+12025550123", normalized)

	// Invalid empty
	_, err = utils.IsValidPhone("")
	require.Error(t, err)

	// Invalid number
	_, err = utils.IsValidPhone("12345")
	require.Error(t, err)
}

func TestEmailHandling(t *testing.T) {
	// With display name
	normalized, err := utils.IsValidEmail("John Doe <john@example.com>")
	require.NoError(t, err)
	assert.Equal(t, "john@example.com", normalized)

	// Gmail subaddressing stripped
	assert.Equal(t, "user@gmail.com", utils.NormalizeEmail("user+promo@gmail.com"))
	assert.Equal(t, "user@googlemail.com", utils.NormalizeEmail("user+alerts@googlemail.com"))

	// Non-gmail subaddressing preserved
	assert.Equal(t, "user+promo@customdomain.com", utils.NormalizeEmail("user+promo@customdomain.com"))
}

func TestSlugify(t *testing.T) {
	assert.Equal(t, "at-and-t", utils.Slugify("AT&T"))
	assert.Equal(t, "c-plus-plus", utils.Slugify("C++"))
	assert.Equal(t, "a-b", utils.Slugify("A/B"))
	assert.Equal(t, "cafe-creme", utils.Slugify("Café Crème"))
	assert.Equal(t, "hello-world", utils.Slugify("Hello World!"))
	assert.Equal(t, "n-a", utils.Slugify("???"))
	assert.Equal(t, "custom-fallback", utils.Slugify("???", "custom-fallback"))
	assert.Equal(t, "product-title", utils.Slugify("--- Product Title ---"))
}
