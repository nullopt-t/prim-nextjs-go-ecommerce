const url = import.meta.env.VITE_BASE_URL;

async function PostData(endpoint, payload) {
	try {
		const response = await fetch(`${url}${endpoint}`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(payload),
		});
		if (!response.ok) throw Error("HTTP ERROR.");
		const data = await response.json();
		return data;
	} catch (err) {
		throw new Error("Failed", { cause: err });
	}
}

export default PostData;
