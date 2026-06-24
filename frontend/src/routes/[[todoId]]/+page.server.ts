import { TOKEN_COOKIE_NAME } from '$lib/constants';
import type { Todo } from '$lib/todo/todo';
import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ cookies }) => {
	if (cookies.get(TOKEN_COOKIE_NAME) === undefined) {
		error(403, 'Please log in');
	}

	const testTodo: Todo = {
		name: 'Project Planning',
		description: 'Plan the next project phase.\nhello\n\ngoodbye',
		children: [
			{
				id: 1,
				name: 'Define Scope',
				description: 'Clearly define the project boundaries.'.repeat(100),
				valueRank: 5,
				timeRank: 3,
			},
			{
				id: 2,
				name: 'Assign Tasks',
				description: 'Distribute tasks among team members.',
				valueRank: 4,
				timeRank: 1,
			},
			{
				id: 3,
				name: 'Set Deadlines',
				description: 'Establish realistic deadlines for each task.',
				valueRank: 3,
				timeRank: 2,
			},
			{
				id: 4,
				name: 'Resource Allocation',
				description: 'Allocate necessary resources.',
				valueRank: 2,
				timeRank: 4,
			},
			{
				id: 5,
				name: 'Risk Assessment',
				description: 'Identify and assess potential risks.',
				valueRank: 1,
				timeRank: 5,
			},
		],
		parentId: null,
	};

	return { todo: testTodo };
};
