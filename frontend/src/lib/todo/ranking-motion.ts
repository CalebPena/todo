export const rankingMotion = {
	duration: 1000,
	easing: 'cubic-bezier(0.22, 1, 0.36, 1)',
};

export function rankingEase(progress: number) {
	if (progress <= 0 || progress >= 1) return progress;
	let low = 0;
	let high = 1;
	for (let index = 0; index < 18; index++) {
		const t = (low + high) / 2;
		const x = 3 * (1 - t) ** 2 * t * 0.22 + 3 * (1 - t) * t ** 2 * 0.36 + t ** 3;
		if (x < progress) low = t;
		else high = t;
	}
	return 1 - (1 - (low + high) / 2) ** 3;
}
