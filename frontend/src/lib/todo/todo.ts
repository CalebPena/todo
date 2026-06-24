export type SubTodo = {
	id: number;
	name: string;
	description: string;
	valueRank: number;
	timeRank: number;
};

export type Todo = {
	name: string;
	description: string;
	children: SubTodo[];
	parentId: number | null;
};

export function todoPriority(todo: SubTodo, todoCount: number): number {
	return todoCount - todo.timeRank + todo.valueRank;
}
