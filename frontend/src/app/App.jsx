import Router from "@/app/router";
import useDirection from "@/hooks/useDirection";

function App() {
	useDirection();

	return <Router />;
}

export default App;
