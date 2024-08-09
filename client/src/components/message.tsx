import { ArrowUp } from "lucide-react";
import { useState } from "react";

interface MessageProps {
	text: string;
	amountOfReactions: number;
	answered?: boolean;
}

export function Message({
	text,
	amountOfReactions,
	answered = false,
}: MessageProps) {
	const [hasReacted, setHasReacted] = useState(false);
	const [reactions, setReactions] = useState(0);

	function handleReactToMessage() {
		setHasReacted(true);
		setReactions(reactions + 1);
	}
	function handleRemoveReactFromMessage() {
		setHasReacted(false);
		setReactions(reactions - 1);
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
					<ArrowUp /> Curtir ({amountOfReactions})
				</button>
			) : (
				<button
					type="button"
					onClick={handleReactToMessage}
					className="mt-3 flex items-center gap-2 text-zinc-400 text-sm font-medium hover:text-zinc-300"
				>
					<ArrowUp /> Curtir ({amountOfReactions})
				</button>
			)}
		</li>
	);
}
