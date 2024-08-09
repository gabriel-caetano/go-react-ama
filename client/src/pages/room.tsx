import { Share2 } from "lucide-react";
import { Suspense } from "react";
import { useParams } from "react-router-dom";
import { toast } from "sonner";
import amaLogo from "../assets/ama-logo.svg";
import { CreateMessageForm } from "../components/create-message-form";
import { Messages } from "../components/messages";

export function Room() {
	const { roomId } = useParams();

	function handleShareRoom() {
		const url = window.location.href.toString();
		if (navigator.share !== undefined && navigator.canShare()) {
			navigator.share({ url });
		} else {
			navigator.clipboard.writeText(url);
			toast.info("O link da sala foi copiado para a area de transferência");
		}
	}

	return (
		<div className="mx-auto max-w-[640px] flex flex-col gap-6 py-10 px-4">
			<div className="flex items-center gap-3 px-3">
				<img src={amaLogo} alt="AMA logo" className="h-5" />
				<span className="text-sm text-zinc-500 truncate">
					Código da sala: <span className="text-zinc-300">{roomId}</span>
				</span>
				<button
					type="submit"
					onClick={handleShareRoom}
					className="ml-auto bg-zinc-800 text-zinc-300 px-3 py-1.5 flex items-center rounded-lg font-medium text-sm hover:bg-zinc-700 transition-colors"
				>
					Compartilhar
					<Share2 className="size-4 pl-1" />
				</button>
			</div>
			<div className="h-px w-full bg-zinc-900" />
			<CreateMessageForm />
			<Suspense fallback={<p>Carregando...</p>}>
				<Messages />
			</Suspense>
		</div>
	);
}
