import { useQueryClient, useSuspenseQuery } from "@tanstack/react-query";
import { useEffect } from "react";
import { useParams } from "react-router-dom";
import {
	type GetRoomMessagesReturn,
	getRoomMessages,
} from "../http/get-room-messages";
import { getMessage } from "../tools/get-message-from-ws";
import { Message } from "./message";

export function Messages() {
	const { roomId } = useParams();
	const queryClient = useQueryClient();

	if (!roomId) {
		throw new Error("Messages components must be used within room page");
	}

	const { data } = useSuspenseQuery({
		queryKey: ["messages", roomId],
		queryFn: () => getRoomMessages({ roomId }),
	});

	useEffect(() => {
		const ws = new WebSocket(`ws://localhost:8080/subscribe/${roomId}`);

		ws.onopen = () => {
			console.log("websocket connected");
		};

		ws.onmessage = (e) => {
			const data = JSON.parse(e.data);

			switch (data.kind) {
				case "message_created":
					console.log("updating messages...");

					console.log({ m: getMessage(data.value) });
					queryClient.setQueryData<GetRoomMessagesReturn>(
						["messages", roomId],
						(state) => {
							const messages = [
								...(state?.messages ?? []),
								getMessage(data.value),
							];
							console.log(messages);

							return {
								messages,
							};
						},
					);
			}
		};

		return () => {
			ws.close();
		};
	}, [roomId, queryClient]);

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
