import { useSuspenseQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { getRoomMessages } from "../http/get-room-messages";
import { Message } from "./message";

export function Messages() {
	const { roomId } = useParams();

	if (!roomId) {
		throw new Error("Messages components must be used within room page");
	}

	const { data } = useSuspenseQuery({
		queryKey: ["messages", roomId],
		queryFn: () => getRoomMessages({ roomId }),
	});

	return (
		<ol className="list-decimal list-inside px-3 space-y-8">
			{data.messages.map((message) => {
				return (
					<Message
						id={message.id}
						key={message.id}
						answered={message.answered}
						text={message.text}
						amountOfReactions={message.amountOfReactions}
					/>
				);
			})}
		</ol>
	);
}
