import type { MessageData } from "../http/get-room-messages";

interface WebsocketMessageData {
	id: string;
	message: string;
	reaction_count: number;
	answered: boolean;
}

export function getMessage(value: WebsocketMessageData): MessageData {
	return {
		id: value.id,
		text: value.message,
		amountOfReactions: value.reaction_count,
		answered: value.answered,
	};
}
