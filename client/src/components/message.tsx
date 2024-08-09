import { ArrowUp } from "lucide-react";
import { useState } from "react";
import { useParams } from "react-router-dom";
import { toast } from "sonner";
import { createMessageReaction } from "../http/create-message-reaction";
import { removeMessageReaction } from "../http/remove-message-reaction";

interface MessageProps {
	id: string;
	text: string;
	amountOfReactions: number;
	answered?: boolean;
}

export function Message({
	id: messageId,
	text,
	amountOfReactions,
	answered = false,
}: MessageProps) {
	const { roomId } = useParams();
	if (!roomId) {
		throw new Error("Message components must be used within room page");
	}
	const [hasReacted, setHasReacted] = useState(false);
	const [reactions, setReactions] = useState(amountOfReactions);

	async function handleReactToMessage() {
		if (!roomId) {
			return;
		}
		try {
			const data = await createMessageReaction({ roomId, messageId });
			setReactions(data.amountOfReactions);
			setHasReacted(true);
		} catch (e) {
			console.log(e);

			toast.error("Falha ao curtir mensagem, tente novamente");
		}
	}
	async function handleRemoveReactFromMessage() {
		if (!roomId) {
			return;
		}
		try {
			const data = await removeMessageReaction({ roomId, messageId });
			setReactions(data.amountOfReactions);
			setHasReacted(false);
		} catch (e) {
			toast.error("Falha ao remover curtida mensagem, tente novamente");
			console.log(e);
		}
	}

	return (
		<li
			data-answered={answered}
			className="ml-6 leading-relaxed data-[answered=true]:opacity-50 data-[answered=true]:pointer-events-none"
		>
			{text}
			{hasReacted ? (
				<button
					type="button"
					onClick={handleRemoveReactFromMessage}
					className="mt-3 flex items-center gap-2 text-orange-400 text-sm font-medium hover:text-orange-500"
				>
					<ArrowUp /> Curtir ({reactions})
				</button>
			) : (
				<button
					type="button"
					onClick={handleReactToMessage}
					className="mt-3 flex items-center gap-2 text-zinc-400 text-sm font-medium hover:text-zinc-300"
				>
					<ArrowUp /> Curtir ({reactions})
				</button>
			)}
		</li>
	);
}
