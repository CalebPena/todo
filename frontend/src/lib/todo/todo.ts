import { z } from 'zod';

const todoFields = z.object({
	id: z.number().int().positive(),
	name: z.string(),
	description: z.string(),
	completed: z.boolean(),
	ranked: z.boolean(),
	valueRank: z.number().int().positive().nullable(),
	timeRank: z.number().int().positive().nullable(),
	parentId: z.number().int().positive().nullable(),
});

export type Todo = z.infer<typeof todoFields> & { children: Todo[] };

export const todoSchema: z.ZodType<Todo> = z.lazy(() =>
	todoFields.extend({ children: z.array(todoSchema) }),
);

export const todoInputSchema = z.object({
	name: z
		.string()
		.trim()
		.min(1, 'Please enter a title')
		.refine(
			(name) => new TextEncoder().encode(name).length <= 255,
			'Title must be at most 255 bytes',
		),
	description: z.string(),
});
