package utils

import (
	"errors"
	"math"

	apiresource "k8s.io/apimachinery/pkg/api/resource"
)

// Quantity Multiplication
func QuantityMultiplicative(value apiresource.Quantity, multiplier float64) (result apiresource.Quantity, err error) {
	if value.Format == apiresource.BinarySI {
		quantityValue := int64(multiplier*float64(value.Value())/math.Pow(2, 20)) * int64(math.Pow(2, 20))
		result = *apiresource.NewQuantity(quantityValue, apiresource.BinarySI)
		return
	} else if value.Format == apiresource.DecimalSI {
		quantityValue := int64(multiplier * float64(value.MilliValue()))
		result = *apiresource.NewMilliQuantity(quantityValue, apiresource.DecimalSI)
		return
	} else if value.Format == apiresource.DecimalExponent {
		quantityValue := int64(multiplier * float64(value.Value()))
		result = *apiresource.NewQuantity(quantityValue, apiresource.DecimalExponent)
		return
	} else {
		err = errors.New("resource.Quantity Format is invalid, only support DecimalExponent, DecimalSI, BinarySI")
		return
	}
}
