import { useState, useEffect } from "react";

function useFetch(func, endpoint) {
	const [data, setData] = useState();
	const [errorMsg, setErrorMsg] = useState();
	const [loading, setLoading] = useState(() => true);

	useEffect(() => {
		const brngData = async () => {
			try {
				const result = await func(endpoint);
				setData(result);
			} catch (error) {
				setErrorMsg(error);
			} finally {
				setLoading(false);
			}
		};

		brngData();
	}, [func, endpoint]);

	return { data, loading, errorMsg };
}

export default useFetch;
