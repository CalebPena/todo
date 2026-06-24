import { DEFUALT_ERROR_MESSAGE } from '$lib/constants';
import type { z } from 'zod';

export function formInputError<T>(schema: z.Schema<T>, value: unknown): string {
	const data = schema.safeParse(value);

	if (!data.success) {
		if (data.error.issues.length === 0) {
			return DEFUALT_ERROR_MESSAGE;
		}

		return data.error.issues[0].message;
	}

	return '';
}
