interface CreateRoomRequest {
	theme: string;
}

export async function createRoom({ theme }: CreateRoomRequest) {
	const response = await fetch(`${import.meta.env.BASE_URL}/rooms`, {
		method: "POST",
		body: JSON.stringify({ theme }),
	});

	const data: { id: string } = await response.json();

	return { roomId: data.id };
}
