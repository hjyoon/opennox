int nox_float2int(float a1) {
	// Original419A70 returns integer-indefinite for invalid signed-DWORD input.
	// Native ARM64 conversion otherwise saturates or returns zero for NaN.
	if (!(a1 >= -2147483648.0f && a1 < 2147483648.0f)) {
		return INT32_MIN;
	}
	return (int)a1;
}

float nox_double2float(double a1) {
	// Original419AF0 stores binary32 under the startup ToZero context.
	// Correct a rounded-away native cast without changing host FP state.
	volatile float stored = (float)a1;
	if (fabs((double)stored) > fabs(a1)) {
		return nextafterf(stored, 0);
	}
	return stored;
}

// POINT_SEGMENT_PROJECTION_ORIGINAL427DF0_BEGIN
// The original uses precision53/ToZero, with explicit binary32 stores.
// Int32 geometry bounds all finite intermediates and FMA residuals away
// from binary64 overflow/underflow. Volatile stores prevent contraction.
static double nox_projection_add_chop53_427DF0(double a, double b) {
	volatile double sum = a + b;
	volatile double b_virtual = sum - a;
	volatile double a_virtual = sum - b_virtual;
	volatile double a_residual = a - a_virtual;
	volatile double b_residual = b - b_virtual;
	double residual = a_residual + b_residual;
	if ((sum > 0 && residual < 0) || (sum < 0 && residual > 0)) {
		return nextafter(sum, 0);
	}
	return sum;
}

static double nox_projection_mul_chop53_427DF0(double a, double b) {
	volatile double product = a * b;
	double residual = fma(a, b, -product);
	if ((product > 0 && residual < 0) || (product < 0 && residual > 0)) {
		return nextafter(product, 0);
	}
	return product;
}

static double nox_projection_div_chop53_427DF0(double a, double b) {
	volatile double quotient = a / b;
	double residual = fma(-quotient, b, a);
	if (residual != 0 && !isnan(residual) &&
		((quotient > 0 && signbit(residual) != signbit(b)) ||
		 (quotient < 0 && signbit(residual) == signbit(b)))) {
		return nextafter(quotient, 0);
	}
	return quotient;
}

static double nox_projection_sqrt_chop53_427DF0(double value) {
	volatile double root = sqrt(value);
	if (fma(root, root, -value) > 0) {
		return nextafter(root, 0);
	}
	return root;
}

static float nox_projection_spill_chop32_427DF0(double value) {
	volatile float stored = (float)value;
	if (fabs((double)stored) > fabs(value)) {
		return nextafterf(stored, 0);
	}
	return stored;
}

//----- (00427DF0) --------------------------------------------------------
int sub_427DF0(int2* point, int* a2, float a3) {
	float x0 = nox_projection_spill_chop32_427DF0((double)a2[0]);
	float y0 = nox_projection_spill_chop32_427DF0((double)a2[1]);
	float dx = nox_projection_spill_chop32_427DF0(
		nox_projection_add_chop53_427DF0((double)a2[2], -(double)x0));
	double dy_live = nox_projection_add_chop53_427DF0((double)a2[3], -(double)y0);
	float dy = nox_projection_spill_chop32_427DF0(dy_live);
	double square = nox_projection_add_chop53_427DF0(
		nox_projection_mul_chop53_427DF0(dy_live, (double)dy),
		nox_projection_mul_chop53_427DF0((double)dx, (double)dx));
	float length = nox_projection_spill_chop32_427DF0(
		nox_projection_sqrt_chop53_427DF0(square));
	float nx = nox_projection_spill_chop32_427DF0(
		nox_projection_div_chop53_427DF0((double)dx, (double)length));
	double ny_live = nox_projection_div_chop53_427DF0((double)dy, (double)length);
	float ny = nox_projection_spill_chop32_427DF0(ny_live);
	float px = nox_projection_spill_chop32_427DF0(
		nox_projection_add_chop53_427DF0((double)point->field_0, -(double)x0));
	float py = nox_projection_spill_chop32_427DF0(
		nox_projection_add_chop53_427DF0((double)point->field_4, -(double)y0));
	double perpendicular = nox_projection_add_chop53_427DF0(
		nox_projection_mul_chop53_427DF0((double)px, -ny_live),
		nox_projection_mul_chop53_427DF0((double)py, (double)nx));
	float distance = nox_projection_spill_chop32_427DF0(fabs(perpendicular));
	if (distance > a3) {
		return 0;
	}
	double projection = nox_projection_add_chop53_427DF0(
		nox_projection_mul_chop53_427DF0((double)py, (double)ny),
		nox_projection_mul_chop53_427DF0((double)px, (double)nx));
	float stored_projection = nox_projection_spill_chop32_427DF0(projection);
	if (projection > (double)length) {
		return 0;
	}
	// Original FCOMP/J(C0) rejects unordered projection as well as negative.
	if (!(stored_projection >= 0)) {
		return 0;
	}
	float out_x = nox_projection_spill_chop32_427DF0(nox_projection_add_chop53_427DF0(
		nox_projection_mul_chop53_427DF0((double)stored_projection, (double)nx), (double)x0));
	float out_y = nox_projection_spill_chop32_427DF0(nox_projection_add_chop53_427DF0(
		nox_projection_mul_chop53_427DF0((double)stored_projection, (double)ny), (double)y0));
	point->field_0 = nox_float2int(out_x);
	point->field_4 = nox_float2int(out_y);
	return 1;
}
