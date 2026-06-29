import { useState, useEffect } from "react";
import GetData from "@/api/get";

function useFetch(endpoint) {
	const [data, setData] = useState();
	const [errorMsg, setErrorMsg] = useState();
	const [loading, setLoading] = useState(() => true);

	useEffect(() => {
		const brngData = async () => {
			try {
				const result = await GetData(endpoint);
				setData(result);
			} catch (error) {
				setErrorMsg(error);
			} finally {
				setLoading(false);
			}
		};

		brngData();
	}, [endpoint]);

	return { data, loading, errorMsg };
}

export default useFetch;
