import type { CardPose, RankingClip, RankingScene } from './ranking-scene';
import { rankingMotion } from './ranking-motion';

export type RankingCardElements = { root: HTMLElement; flipper: HTMLElement; title: HTMLElement };

const transform = (pose: CardPose) =>
	`translate3d(${pose.x}px, ${pose.y}px, 0) rotate(${pose.rotation}deg)`;

export function createRankingTimeline(
	cards: Map<string, RankingCardElements>,
	getExperience: () => HTMLElement | undefined,
) {
	let generation = 0;
	let playing: Animation[] = [];
	let speed = 1;

	function setSpeed(rate: number) {
		speed = Math.max(1, Math.min(4, rate));
		for (const animation of playing) animation.updatePlaybackRate(speed);
	}

	function cancel() {
		generation++;
		for (const animation of playing) animation.cancel();
		playing = [];
	}

	function set(scene: RankingScene) {
		for (const [key, pose] of scene) {
			const card = cards.get(key);
			if (!card) continue;
			Object.assign(card.root.style, {
				visibility: 'visible',
				transform: transform(pose),
				width: `${pose.width}px`,
				height: `${pose.height}px`,
				opacity: String(pose.opacity),
				zIndex: String(pose.z),
			});
			card.flipper.style.transform = `rotateY(${pose.flip}deg)`;
			card.title.style.fontSize = `${pose.fontSize}px`;
		}
	}

	async function play(clip: RankingClip, reverse: boolean, reducedMotion: boolean) {
		cancel();
		const current = generation;
		const destination = reverse ? clip.start : clip.end;
		set(reverse ? clip.end : clip.start);
		if (reducedMotion || clip.tracks.length === 0) {
			set(destination);
			return true;
		}
		const options: KeyframeAnimationOptions = {
			duration: clip.duration,
			easing: 'linear',
			fill: 'both',
		};
		for (const track of clip.tracks) {
			const card = cards.get(track.key);
			if (!card) continue;
			const frames = reverse ? track.reverseFrames : track.frames;
			const rootFrames = frames.map(({ offset, pose }) => ({
				offset,
				transform: transform(pose),
				width: `${pose.width}px`,
				height: `${pose.height}px`,
				opacity: pose.opacity,
				zIndex: pose.z,
			}));
			playing.push(
				card.root.animate(rootFrames, options),
				card.flipper.animate(
					frames.map(({ offset, pose }) => ({
						offset,
						transform: `rotateY(${pose.flip}deg)`,
					})),
					options,
				),
				card.title.animate(
					frames.map(({ offset, pose }) => ({ offset, fontSize: `${pose.fontSize}px` })),
					options,
				),
			);
		}
		const experience = getExperience();
		if (experience && clip.accent.from !== clip.accent.to) {
			const styles = getComputedStyle(experience);
			const from = styles
				.getPropertyValue(`--ranking-${reverse ? clip.accent.to : clip.accent.from}-color`)
				.trim();
			const to = styles
				.getPropertyValue(`--ranking-${reverse ? clip.accent.from : clip.accent.to}-color`)
				.trim();
			playing.push(
				experience.animate(
					[
						{ '--ranking-choice-color': from, easing: rankingMotion.easing },
						{ '--ranking-choice-color': to },
					],
					options,
				),
			);
		}
		for (const animation of playing) {
			animation.pause();
			animation.playbackRate = speed;
			animation.currentTime = 0;
		}
		const startTime = document.timeline.currentTime;
		for (const animation of playing) {
			animation.play();
			if (startTime !== null) animation.startTime = startTime;
		}
		await Promise.allSettled(playing.map((animation) => animation.finished));
		if (current !== generation) return false;
		set(destination);
		for (const animation of playing) animation.cancel();
		playing = [];
		return true;
	}

	return { set, play, cancel, setSpeed };
}
