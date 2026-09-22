import { Link, useRouter, usePathname } from "@/i18n/navigation"; import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import MainLayout from "@/components/layouts/mainLayout";
import { LazyImage } from "@/components/ui/lazyImage";
import { Stars } from "@/components/ui/stars";
import { AnimatedSection, CustomButton } from "@/components/ui";
import { 
	Heart, 
	ShoppingCart, 
	ChevronRight, 
	Share2, 
	Truck, 
	ShieldCheck, 
	Check, 
	Clock, 
	Package, 
	User, 
	ThumbsUp,
	Zap
} from "lucide-react";
import { useState, useMemo, useEffect } from "react";
import QuantitySelector from "@/components/ui/quantitySelector";
import { ProductsGrid } from "@/components/ui/productsGrid";
import SectionTitle from "@/features/home/components/ui/sectionTitle";
import { useAllProducts, useProductBySlug } from "@/hooks/useCatalog";
import { useCartContext } from "@/context/CartContext";
import { catalogService } from "@/services/catalog";
import { api } from "@/api/client";
import { toast } from "sonner";

export function ProductDetails() {
	const { id } = useParams();
	const navigate = useRouter();
	const t = useTranslations("common");
	
	const { products: allProducts, loading: allLoading } = useAllProducts();
	const { product, loading: productLoading } = useProductBySlug(id);
	const { addToCart, openCart } = useCartContext();

	const [activeImage, setActiveImage] = useState(0);
	const [selectedColor, setSelectedColor] = useState(0);
	const [quantity, setQuantity] = useState(1);
	const [isAdding, setIsAdding] = useState(false);
	const [isWishlisted, setIsWishlisted] = useState(false);
	const [isReviewFormOpen, setIsReviewFormOpen] = useState(false);
	const [reviewForm, setReviewForm] = useState({ rating: 0, title: "", content: "" });
	const [isReviewsExpanded, setIsReviewsExpanded] = useState(false);

	useEffect(() => {
		window.scrollTo(0, 0);
		// setActiveImage(0);
		// setSelectedColor(0);
	}, [id]);

	const [reviews, setReviews] = useState<any[]>([]);
	const [ratingSummary, setRatingSummary] = useState<any>(null);

	useEffect(() => {
		let isMounted = true;
		const cleanSlug = Array.isArray(id) ? id[0] : id;
		if (!cleanSlug) return;

		Promise.allSettled([
			catalogService.getProductReviews(cleanSlug),
			catalogService.getProductRatingSummary(cleanSlug),
		]).then(([revRes, sumRes]) => {
			if (!isMounted) return;
			if (revRes.status === "fulfilled" && revRes.value) {
				const list = Array.isArray(revRes.value) 
					? revRes.value 
					: (Array.isArray(revRes.value?.data) ? revRes.value.data : []);
				if (list.length > 0) {
					const mapped = list.map((r: any) => ({
						id: r.id,
						author: r.user?.name || "Customer",
						avatar: r.user?.avatar || null,
						rating: r.rating || 5,
						date: r.createdAt ? new Date(r.createdAt).toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" }) : "Recently",
						title: r.title || "Customer Review",
						content: r.body || r.content || "",
						helpful: r.helpful || 0,
						verified: true,
					}));
					setReviews(mapped);
				}
			}
			if (sumRes.status === "fulfilled" && sumRes.value) {
				setRatingSummary(sumRes.value?.data || sumRes.value);
			}
		});

		return () => {
			isMounted = false;
		};
	}, [id]);

	const similarProducts = useMemo(() => {
		if (!product || !allProducts) return [];
		return allProducts
			.filter(p => p.category === product.category && p.id !== product.id)
			.slice(0, 5);
	}, [product, allProducts]);

	if (allLoading || productLoading) {
		return (
			<div className="flex flex-col items-center justify-center py-32 text-center text-muted-foreground">
				Loading...
			</div>
		);
	}

	if (!product) {
		return (
			<div className="flex flex-col items-center justify-center py-32 text-center">
				<Package className="size-16 text-muted-foreground mb-4 opacity-50" />
				<h1 className="text-2xl font-bold text-foreground mb-4">Product Not Found</h1>
				<p className="text-muted-foreground mb-8">The product you are looking for doesn't exist or has been removed.</p>
				<Link href="/products">
					<CustomButton text="Browse All Products" className="px-8" />
				</Link>
			</div>
		);
	}

	const productName = typeof product.product === "object"
		? (product.product?.en || product.product?.ar || "Product")
		: (product.product || product.name || "Product");

	const handleAddToCart = async () => {
		if (!(product.inStock ?? true)) return;
		setIsAdding(true);
		const defaultVariant = (product.variants && product.variants[0]) || null;
		const payload = {
			id: String(product.id),
			variantId: defaultVariant?.id || String(product.id),
			productName: product.product,
			productPrice: `$${product.price}`,
			img: (product.images || [product.img])[0],
			quantity,
			color: (product.colors || [])[selectedColor]?.name || "Default",
		};
		const res = await addToCart(payload);
		setIsAdding(false);
		if (res.success) {
			toast.success(`Added ${quantity}x ${productName} to your cart!`);
			openCart();
		} else {
			toast.error(res.error || "Failed to add to cart");
		}
	};

	const handleBuyNow = async () => {
		if (!(product.inStock ?? true)) return;
		setIsAdding(true);
		const payload = {
			id: String(product.id),
			productName: product.product,
			productPrice: `$${product.price}`,
			img: (product.images || [product.img])[0],
			quantity,
			color: (product.colors || [])[selectedColor]?.name || "Default",
		};
		const res = await addToCart(payload);
		setIsAdding(false);
		if (res.success) {
			navigate.push("/checkout");
		} else {
			toast.error(res.error || "Failed to initiate checkout");
		}
	};

	const handleToggleWishlist = async () => {
		try {
			if (!isWishlisted) {
				await api.post("/api/v1/wishlist", {
					id: String(product.id),
					productName: product.product,
					productPrice: `$${product.price}`,
					img: (product.images || [product.img])[0],
				});
				setIsWishlisted(true);
				toast.success(`Saved ${productName} to your wishlist!`);
			} else {
				await api.delete(`/api/v1/wishlist/${product.id}`);
				setIsWishlisted(false);
				toast.info(`Removed ${productName} from wishlist`);
			}
		} catch {
			toast.error("Failed to update wishlist");
		}
	};

	const handleShare = () => {
		if (navigator.clipboard) {
			navigator.clipboard.writeText(window.location.href);
			toast.success("Product link copied to clipboard!");
		} else {
			toast.info("Share URL: " + window.location.href);
		}
	};

	return (
		<AnimatedSection className="max-w-7xl mx-auto">
				{/* Breadcrumb */}
				<nav className="flex items-center gap-2 text-sm text-muted-foreground mb-8 overflow-x-auto whitespace-nowrap pb-2 scrollbar-none">
					<Link href="/" className="hover:text-foreground transition-colors">Home</Link>
					<ChevronRight className="size-4 rtl:rotate-180 shrink-0" />
					<Link href="/products" className="hover:text-foreground transition-colors">Products</Link>
					<ChevronRight className="size-4 rtl:rotate-180 shrink-0" />
					<Link 
						href={`/products?category=${product.categoryId || product.category}`} 
						className="hover:text-foreground transition-colors capitalize"
					>
						{product.category?.replace?.("-", " ") || product.category}
					</Link>
					<ChevronRight className="size-4 rtl:rotate-180 shrink-0" />
					<span className="text-foreground font-medium truncate max-w-[200px]">{productName}</span>
				</nav>

				<div className="grid grid-cols-1 md:grid-cols-2 gap-10 lg:gap-16">
					{/* Image Gallery */}
					<div className="flex flex-col gap-4">
						<div className="relative aspect-square w-full overflow-hidden rounded-2xl border border-border/50 bg-secondary/30">
							<LazyImage
								src={(product.images || [product.img])[activeImage]}
								alt={productName}
								containerClassName="w-full h-full"
								imageClassName="object-cover object-center w-full h-full"
							/>
							{product.discountPercentage && (
								<span className="absolute top-4 left-4 bg-destructive text-white px-3 py-1 rounded-full text-sm font-semibold shadow-sm">
									{product.discountPercentage} OFF
								</span>
							)}
						</div>
						
						{/* Thumbnails */}
						<div className="grid grid-cols-4 gap-4">
							{(product.images || [product.img]).map((img, i) => (
								<button 
									key={i} 
									onClick={() => setActiveImage(i)}
									className={`aspect-square rounded-xl overflow-hidden border ${i === activeImage ? 'border-accent-brand ring-2 ring-accent-brand/20' : 'border-border/50 hover:border-border'} transition-all`}
								>
									<LazyImage
										src={img}
										alt={`${productName} thumbnail ${i + 1}`}
										containerClassName={`w-full h-full transition-opacity ${i === activeImage ? 'opacity-100' : 'opacity-70 hover:opacity-100'}`}
										imageClassName="object-cover object-center w-full h-full"
									/>
								</button>
							))}
						</div>
					</div>

					{/* Product Info */}
					<div className="flex flex-col">
						<div className="flex flex-col gap-2 mb-4">
							<span className="text-sm font-medium text-accent-brand capitalize">{product.category.replace("-", " ")}</span>
							<h1 className="text-3xl lg:text-4xl font-bold text-foreground">{productName}</h1>
						</div>
						
						<div className="flex items-center gap-4 mb-6">
							<div className="flex items-center gap-1.5">
								<Stars starsNum={product.stars} />
								<span className="text-sm font-medium text-foreground ml-1">{product.stars}</span>
							</div>
							<div className="w-1 h-1 rounded-full bg-border" />
							<span className="text-sm text-muted-foreground underline decoration-dashed cursor-pointer hover:text-foreground transition-colors">
								{product.reviews} Reviews
							</span>
						</div>

						<div className="flex items-end gap-3 mb-6">
							<span className="text-4xl font-bold text-foreground">
								{t("currency")}{product.price}
							</span>
							{product.oldPrice && (
								<span className="text-xl text-muted-foreground line-through mb-1">
									{t("currency")}{product.oldPrice}
								</span>
							)}
						</div>

						<p className="text-muted-foreground text-base leading-relaxed mb-6">
							We've carefully designed the {productName} to be something you'll love using every single day. It's built to last, feels great to use, and fits right into your lifestyle without any fuss. Whether you're picking it up for the first time or the hundredth, we think you'll appreciate the little details we've put into it.
						</p>

						{/* Highlights */}
						<div className="mb-8">
							<h3 className="text-foreground font-semibold mb-3">Product Highlights</h3>
							<ul className="list-disc pl-5 space-y-1.5 text-sm text-muted-foreground">
								<li>Feels great to hold and use for hours</li>
								<li>Built sturdy to handle everyday bumps and scratches</li>
								<li>Super easy to set up and get started</li>
								<li>Packaged thoughtfully without unnecessary plastics</li>
							</ul>
						</div>

						{/* Variants - Colors */}
						{(product.colors || []).length > 0 && (
							<div className="flex flex-col gap-3 mb-8">
								<span className="text-sm font-medium text-foreground">
									Color: <span className="text-muted-foreground ml-1">{(product.colors || [])[selectedColor]?.name || ""}</span>
								</span>
								<div className="flex gap-3">
									{(product.colors || []).map((color: any, idx: number) => (
										<button
											key={idx}
											onClick={() => setSelectedColor(idx)}
											title={color.name}
											className={`size-10 rounded-full border-2 flex items-center justify-center transition-all ${
												selectedColor === idx ? 'border-foreground ring-2 ring-foreground/10 ring-offset-2 ring-offset-background' : 'border-transparent hover:scale-110'
											}`}
										>
											<span className={`w-full h-full rounded-full ${color.class || 'bg-gray-400'}`} />
										</button>
									))}
								</div>
							</div>
						)}

						{/* Stock Status */}
						<div className="flex items-center gap-2 mb-6">
							{(product.inStock ?? true) ? (
								<>
									<div className="flex items-center justify-center size-5 rounded-full bg-emerald-500/10 text-emerald-500">
										<Check className="size-3.5" />
									</div>
									<span className="text-sm font-medium text-emerald-600 dark:text-emerald-400">
										In Stock ({(product.stockCount || 10)} available)
									</span>
								</>
							) : (
								<>
									<div className="flex items-center justify-center size-5 rounded-full bg-destructive/10 text-destructive">
										<Clock className="size-3.5" />
									</div>
									<span className="text-sm font-medium text-destructive">
										Out of Stock
									</span>
								</>
							)}
						</div>

						{/* Actions */}
						<div className="flex flex-col gap-3.5 mb-8">
							<div className="w-full h-12">
								<QuantitySelector 
									initialValue={quantity} 
									max={(product.stockCount || 10)} 
									disabled={!(product.inStock ?? true)} 
									onChange={(val) => setQuantity(val)}
								/>
							</div>

							<div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
								<CustomButton 
									text={isAdding ? "Adding..." : ((product.inStock ?? true) ? "Add to Cart" : "Out of Stock")}
									icon={<ShoppingCart className="size-4" />}
									onClick={handleAddToCart}
									className={`py-3.5 text-sm shadow-sm ${!(product.inStock ?? true) ? 'opacity-50 cursor-not-allowed' : ''}`}
									disabled={isAdding || !(product.inStock ?? true)}
								/>

								<button
									onClick={handleBuyNow}
									disabled={isAdding || !(product.inStock ?? true)}
									className="py-3.5 px-4 rounded-xl bg-accent-brand hover:bg-accent-brand/90 text-white font-semibold text-sm transition-all shadow-sm flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
								>
									<Zap className="size-4 fill-current" />
									<span>Buy Now</span>
								</button>
							</div>

							<div className="flex gap-3">
								<button 
									onClick={handleToggleWishlist}
									className={`flex-1 py-2.5 px-4 flex items-center justify-center gap-2 rounded-xl border transition-all text-xs font-medium shadow-sm ${
										isWishlisted 
											? 'border-red-500/40 bg-red-500/10 text-red-600' 
											: 'border-border bg-card hover:bg-secondary hover:text-foreground text-muted-foreground'
									}`}
								>
									<Heart className={`size-4 ${isWishlisted ? 'fill-current text-red-600' : ''}`} />
									<span>{isWishlisted ? "Saved to Wishlist" : "Add to Wishlist"}</span>
								</button>

								<button 
									onClick={handleShare}
									className="py-2.5 px-4 flex items-center justify-center gap-2 rounded-xl border border-border bg-card hover:bg-secondary text-muted-foreground hover:text-foreground transition-colors shadow-sm text-xs font-medium"
								>
									<Share2 className="size-4" />
									<span>Share</span>
								</button>
							</div>
						</div>

						{/* Features */}
						<div className="grid grid-cols-1 sm:grid-cols-2 gap-4 py-6 border-y border-border">
							<div className="flex items-center gap-3 text-muted-foreground">
								<div className="p-2 bg-secondary/50 rounded-full text-foreground">
									<Truck className="size-5" />
								</div>
								<div className="flex flex-col">
									<span className="text-sm font-medium text-foreground">Free Shipping</span>
									<span className="text-xs">Worldwide delivery</span>
								</div>
							</div>
							<div className="flex items-center gap-3 text-muted-foreground">
								<div className="p-2 bg-secondary/50 rounded-full text-foreground">
									<ShieldCheck className="size-5" />
								</div>
								<div className="flex flex-col">
									<span className="text-sm font-medium text-foreground">Secure Payment</span>
									<span className="text-xs">100% protected</span>
								</div>
							</div>
						</div>
					</div>
				</div>

				{/* Product Information Sections */}
				<div className="mt-20 flex flex-col gap-16">
					{/* Description Section */}
					<section>
						<h2 className="text-2xl font-bold text-foreground mb-6">Description</h2>
						<div className="prose prose-sm max-w-none text-muted-foreground bg-card border border-border rounded-2xl p-6 md:p-8 shadow-sm">
							<p>
								Enhance your lifestyle with this exceptional product. Built with premium materials 
								and cutting-edge technology, it delivers outstanding performance and reliability. 
								Whether you're a professional or enthusiast, you'll appreciate the attention to 
								detail and superior build quality.
							</p>
							<p className="mt-4">
								Every component has been carefully selected to ensure longevity and maximum satisfaction. 
								The intuitive design means you spend less time figuring it out and more time enjoying it.
							</p>
							<ul className="mt-4 space-y-2">
								<li>High-quality materials for lasting durability</li>
								<li>Ergonomic design for maximum comfort</li>
								<li>Advanced features for enhanced productivity</li>
								<li>Eco-friendly packaging and sustainable production</li>
							</ul>
						</div>
					</section>

					{/* Specifications Section */}
					<section>
						<h2 className="text-2xl font-bold text-foreground mb-6">Specifications</h2>
						<div className="bg-card border border-border rounded-2xl p-6 md:p-8 shadow-sm">
							<div className="grid grid-cols-1 sm:grid-cols-2 gap-x-12 gap-y-4 text-sm">
								<div className="flex justify-between py-3 border-b border-border/50">
									<span className="text-muted-foreground">Brand</span>
									<span className="font-medium text-foreground">{productName.split(' ')[0]}</span>
								</div>
								<div className="flex justify-between py-3 border-b border-border/50">
									<span className="text-muted-foreground">Category</span>
									<span className="font-medium text-foreground capitalize">{product.category.replace("-", " ")}</span>
								</div>
								<div className="flex justify-between py-3 border-b border-border/50">
									<span className="text-muted-foreground">Model Year</span>
									<span className="font-medium text-foreground">2026</span>
								</div>
								<div className="flex justify-between py-3 border-b border-border/50">
									<span className="text-muted-foreground">Condition</span>
									<span className="font-medium text-foreground">Brand New</span>
								</div>
								<div className="flex justify-between py-3 border-b border-border/50">
									<span className="text-muted-foreground">Weight</span>
									<span className="font-medium text-foreground">1.2 kg</span>
								</div>
								<div className="flex justify-between py-3 border-b border-border/50">
									<span className="text-muted-foreground">Dimensions</span>
									<span className="font-medium text-foreground">20 x 15 x 5 cm</span>
								</div>
							</div>
						</div>
					</section>

					{/* Shipping & Returns Section */}
					<section>
						<h2 className="text-2xl font-bold text-foreground mb-6">Shipping & Returns</h2>
						<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div className="bg-card border border-border rounded-2xl p-6 flex flex-col gap-3 shadow-sm hover:border-accent-brand/50 transition-colors">
								<div className="w-10 h-10 rounded-full bg-secondary/50 flex items-center justify-center text-foreground">
									<Truck className="size-5" />
								</div>
								<h4 className="text-foreground font-bold text-base">Shipping Information</h4>
								<p className="text-sm text-muted-foreground leading-relaxed">
									We offer free standard shipping on all orders worldwide. Delivery typically takes 3-5 business days depending on your location. Expedited shipping is available at checkout for an additional fee.
								</p>
							</div>
							
							<div className="bg-card border border-border rounded-2xl p-6 flex flex-col gap-3 shadow-sm hover:border-accent-brand/50 transition-colors">
								<div className="w-10 h-10 rounded-full bg-secondary/50 flex items-center justify-center text-foreground">
									<ShieldCheck className="size-5" />
								</div>
								<h4 className="text-foreground font-bold text-base">Return Policy</h4>
								<p className="text-sm text-muted-foreground leading-relaxed">
									If you're not completely satisfied with your purchase, you can return it within 30 days for a full refund or exchange. The item must be in its original condition and packaging. Please contact our support team to initiate a return.
								</p>
							</div>
						</div>
					</section>

					{/* Reviews Section */}
					<section>
						<h2 className="text-2xl font-bold text-foreground mb-6">Customer Reviews</h2>
						<div className="bg-card border border-border rounded-2xl p-6 md:p-8 shadow-sm">
							<div className="grid grid-cols-1 lg:grid-cols-12 gap-10">
								
								{/* Reviews Summary - Left Column */}
								<div className="lg:col-span-4 flex flex-col">
									<div className="flex items-center gap-4 mb-4">
										<Stars starsNum={product.stars} />
										<span className="text-xl font-bold text-foreground">{product.stars} out of 5</span>
									</div>
									<p className="text-sm text-muted-foreground mb-6">{product.reviews} global ratings</p>
									
									{/* Rating Bars */}
									<div className="flex flex-col gap-2 mb-8">
										{[
											{ stars: 5, pct: 72 },
											{ stars: 4, pct: 18 },
											{ stars: 3, pct: 6 },
											{ stars: 2, pct: 2 },
											{ stars: 1, pct: 2 }
										].map((bar) => (
											<div key={bar.stars} className="flex items-center gap-3 text-sm">
												<span className="w-12 text-accent-brand hover:underline cursor-pointer font-medium whitespace-nowrap">
													{bar.stars} star
												</span>
												<div className="flex-1 h-4 bg-secondary/50 rounded-full overflow-hidden border border-border/50">
													<div 
														className="h-full bg-amber-400 rounded-full"
														style={{ width: `${bar.pct}%` }}
													/>
												</div>
												<span className="w-10 text-right text-muted-foreground">{bar.pct}%</span>
											</div>
										))}
									</div>

									<div className="border-t border-border pt-6">
										<h4 className="font-bold text-foreground mb-2">Review this product</h4>
										<p className="text-sm text-muted-foreground mb-4">Share your thoughts with other customers</p>
										{!isReviewFormOpen ? (
											<button 
												onClick={() => setIsReviewFormOpen(true)}
												className="w-full py-2 px-4 rounded-lg bg-secondary text-foreground hover:bg-secondary/80 border border-border shadow-sm text-sm font-medium transition-colors"
											>
												Write a customer review
											</button>
										) : (
											<div className="flex flex-col gap-4 mt-2 animate-in fade-in slide-in-from-top-2 duration-300">
												<div className="flex flex-col gap-1.5">
													<label className="text-sm font-semibold text-foreground">Rating</label>
													<div className="flex items-center gap-1">
														{[1,2,3,4,5].map((star) => (
															<button 
																key={star}
																onClick={() => setReviewForm({ ...reviewForm, rating: star })}
																className={`text-2xl transition-colors hover:scale-110 ${reviewForm.rating >= star ? "text-amber-400" : "text-border"}`}
															>
																★
															</button>
														))}
													</div>
												</div>
												<div className="flex flex-col gap-1.5">
													<label className="text-sm font-semibold text-foreground">Add a headline</label>
													<input 
														type="text" 
														placeholder="What's most important to know?"
														className="w-full px-3 py-2 text-sm rounded-lg border border-border bg-background focus:outline-none focus:ring-2 focus:ring-accent-brand focus:border-transparent transition-all"
														value={reviewForm.title}
														onChange={(e) => setReviewForm({ ...reviewForm, title: e.target.value })}
													/>
												</div>
												<div className="flex flex-col gap-1.5">
													<label className="text-sm font-semibold text-foreground">Add a written review</label>
													<textarea 
														placeholder="What did you like or dislike? What did you use this product for?"
														className="w-full px-3 py-2 text-sm rounded-lg border border-border bg-background min-h-[120px] resize-y focus:outline-none focus:ring-2 focus:ring-accent-brand focus:border-transparent transition-all"
														value={reviewForm.content}
														onChange={(e) => setReviewForm({ ...reviewForm, content: e.target.value })}
													/>
												</div>
												<div className="flex gap-3 mt-2">
													<button 
														onClick={() => {
															setIsReviewFormOpen(false);
															setReviewForm({ rating: 0, title: "", content: "" });
														}}
														className="flex-1 py-2 px-4 rounded-lg bg-secondary text-foreground hover:bg-secondary/80 border border-border shadow-sm text-sm font-medium transition-colors"
													>
														Cancel
													</button>
													<button 
														onClick={() => {
															setReviews((prev) => [
																{
																	id: Date.now(),
																	author: "You",
																	avatar: null,
																	rating: reviewForm.rating,
																	date: `Reviewed on ${new Date().toLocaleDateString("en-US", { year: 'numeric', month: 'long', day: 'numeric' })}`,
																	title: reviewForm.title,
																	content: reviewForm.content,
																	helpful: 0,
																	verified: true
																},
																...prev
															]);
															setIsReviewFormOpen(false);
															setReviewForm({ rating: 0, title: "", content: "" });
														}}
														className="flex-1 py-2 px-4 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
														disabled={!reviewForm.rating || !reviewForm.title || !reviewForm.content}
													>
														Submit
													</button>
												</div>
											</div>
										)}
									</div>
								</div>

								{/* Reviews List - Right Column */}
								<div className="lg:col-span-8 flex flex-col gap-8">
									{(isReviewsExpanded ? reviews : reviews.slice(0, 2)).map((review) => (
										<div key={review.id} className="flex flex-col border-b border-border/50 pb-8 last:border-0 last:pb-0">
											<div className="flex items-center gap-3 mb-2">
												{review.avatar ? (
													<img src={review.avatar} alt={review.author} className="w-10 h-10 rounded-full object-cover" />
												) : (
													<div className="w-10 h-10 rounded-full bg-secondary flex items-center justify-center text-muted-foreground">
														<User className="size-5" />
													</div>
												)}
												<span className="font-medium text-foreground">{review.author}</span>
											</div>
											
											<div className="flex items-center gap-2 mb-1">
												<Stars starsNum={review.rating} />
												<span className="font-bold text-foreground text-sm">{review.title}</span>
											</div>
											
											<p className="text-xs text-muted-foreground mb-2">{review.date}</p>
											
											{review.verified && (
												<p className="text-xs font-semibold text-accent-brand mb-3">Verified Purchase</p>
											)}
											
											<p className="text-sm text-foreground leading-relaxed mb-4">
												{review.content}
											</p>
											
											<div className="flex items-center gap-4">
												<button 
													onClick={() => {
														setReviews(prev => prev.map(r => {
															if (r.id === review.id) {
																const isVoted = (r as any).userVoted;
																return {
																	...r,
																	helpful: isVoted ? r.helpful - 1 : r.helpful + 1,
																	userVoted: !isVoted
																};
															}
															return r;
														}));
														toast.success((review as any).userVoted ? "Feedback removed" : "Marked as helpful!");
													}}
													className={`flex items-center gap-1.5 text-xs font-medium transition-colors border rounded-full px-3 py-1.5 ${
														(review as any).userVoted 
															? 'border-accent-brand bg-accent-brand/10 text-accent-brand' 
															: 'border-border text-muted-foreground hover:text-foreground hover:bg-secondary'
													}`}
												>
													<ThumbsUp className={`size-3.5 ${(review as any).userVoted ? 'fill-current' : ''}`} />
													Helpful
												</button>
												<span className="text-xs text-muted-foreground">
													{review.helpful} people found this helpful
												</span>
											</div>
										</div>
									))}

									{reviews.length > 2 && (
										<button
											onClick={() => setIsReviewsExpanded(!isReviewsExpanded)}
											className="mt-4 w-full sm:w-auto self-center py-2 px-6 rounded-lg bg-secondary text-foreground hover:bg-secondary/80 border border-border shadow-sm text-sm font-medium transition-colors"
										>
											{isReviewsExpanded ? "Show less" : "Show more"}
										</button>
									)}
								</div>

							</div>
						</div>
					</section>
				</div>

				{/* Similar Products */}
				{similarProducts.length > 0 && (
					<div className="mt-20 mb-12">
						<SectionTitle title="Similar Products" link={`/products?category=${product.categoryId || product.category}`} />
						<ProductsGrid
							products={similarProducts}
							className="grid-cols-2 sm:grid-cols-3 lg:grid-cols-5"
						/>
					</div>
				)}
			</AnimatedSection>
	);
}
