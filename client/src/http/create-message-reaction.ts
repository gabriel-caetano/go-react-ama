interface CreateMessageReactionRequest {
	roomId: string;
	messageId: string;
}

export async function createMessageReaction({
	roomId,
	messageId,
}: CreateMessageReactionRequest) {
	const response = await fetch(
		`${import.meta.env.VITE_APP_API_URL}/rooms/${roomId}/messages/${messageId}/react`,
		{
			method: "PATCH",
		},
	);
	const data: { reaction_count: number } = await response.json();
	return {
		amountOfReactions: data.reaction_count,
	};
}
