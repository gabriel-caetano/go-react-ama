interface GetRoomMessagesRequest {
	roomId: string;
}

export interface MessageData {
	id: string;
	text: string;
	amountOfReactions: number;
	answered: boolean;
}

export interface GetRoomMessagesReturn {
	messages: MessageData[];
}

export async function getRoomMessages({
	roomId,
}: GetRoomMessagesRequest): Promise<GetRoomMessagesReturn> {
	const response = await fetch(
		`${import.meta.env.VITE_APP_API_URL}/rooms/${roomId}/messages`,
	);

	const data: Array<{
		id: string;
		room_id: string;
		message: string;
		reaction_count: string;
		answered: boolean;
	}> = await response.json();
	return {
		messages: data.map((m) => ({
			id: m.id,
			text: m.message,
			amountOfReactions: Number.parseInt(m.reaction_count),
			answered: m.answered,
		})),
	};
}
