export function areExperimentalFeaturesEnabled(): boolean {
	return process.env.TFA_EXPERIMENTAL === "1";
}
