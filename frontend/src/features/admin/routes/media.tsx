import { useState } from "react";
import { UploadCloud, Search, Trash2, Copy, Check, Eye, X } from "lucide-react";
import { INITIAL_MEDIA, getLocalAdminData, setLocalAdminData } from "../data/mockAdminData";
import { toast } from "sonner";

export function AdminMedia() {
  const [mediaList, setMediaList] = useState(() => getLocalAdminData("media", INITIAL_MEDIA));
  const [search, setSearch] = useState("");
  const [copiedId, setCopiedId] = useState(null);
  const [previewItem, setPreviewItem] = useState(null);
  const [isUploading, setIsUploading] = useState(false);

  const saveMedia = (updated) => {
    setMediaList(updated);
    setLocalAdminData("media", updated);
  };

  const handleSimulateUpload = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setIsUploading(true);
    setTimeout(() => {
      const newMedia = {
        id: `med-${Date.now()}`,
        name: file.name,
        size: `${(file.size / (1024 * 1024)).toFixed(1)} MB`,
        type: file.type.includes("image") ? "Image" : "Document",
        dimensions: "1200x1200",
        url: URL.createObjectURL(file) || "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=800&q=80",
        uploadedAt: new Date().toISOString().split("T")[0]
      };
      saveMedia([newMedia, ...mediaList]);
      setIsUploading(false);
      toast.success("File uploaded to media library!");
    }, 800);
  };

  const handleCopyUrl = (url, id) => {
    navigator.clipboard.writeText(url);
    setCopiedId(id);
    toast.success("Media URL copied to clipboard!");
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleDelete = (id) => {
    if (confirm("Delete this media asset?")) {
      saveMedia(mediaList.filter((m) => m.id !== id));
      toast.success("Media file deleted");
    }
  };

  const filtered = mediaList.filter((m) => m.name.toLowerCase().includes(search.toLowerCase()));

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Media Library</h1>
          <p className="text-sm text-muted-foreground mt-1">Upload and manage product catalog images and banners.</p>
        </div>
        <label className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 cursor-pointer transition-all shrink-0">
          <UploadCloud className="size-4" />
          <span>{isUploading ? "Uploading..." : "Upload New File"}</span>
          <input type="file" onChange={handleSimulateUpload} className="hidden" accept="image/*" />
        </label>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs">
        <div className="relative max-w-sm">
          <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search media files..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
          />
        </div>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
        {filtered.map((item) => (
          <div key={item.id} className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs group relative flex flex-col justify-between">
            <div className="aspect-square relative overflow-hidden bg-secondary/20">
              <img src={item.url} alt={item.name} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
              <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                <button onClick={() => setPreviewItem(item)} className="p-2 bg-white/90 text-black rounded-lg hover:bg-white" title="Preview">
                  <Eye className="size-4" />
                </button>
                <button onClick={() => handleCopyUrl(item.url, item.id)} className="p-2 bg-white/90 text-black rounded-lg hover:bg-white" title="Copy Link">
                  {copiedId === item.id ? <Check className="size-4 text-emerald-600" /> : <Copy className="size-4" />}
                </button>
                <button onClick={() => handleDelete(item.id)} className="p-2 bg-white/90 text-red-600 rounded-lg hover:bg-white" title="Delete">
                  <Trash2 className="size-4" />
                </button>
              </div>
            </div>
            <div className="p-3 border-t border-border">
              <div className="font-semibold text-xs text-foreground truncate">{item.name}</div>
              <div className="text-[10px] text-muted-foreground mt-0.5 flex justify-between">
                <span>{item.size}</span>
                <span>{item.dimensions}</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {previewItem && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="font-bold text-sm text-foreground truncate">{previewItem.name}</h3>
              <button onClick={() => setPreviewItem(null)} className="text-muted-foreground hover:text-foreground">
                <X className="size-5" />
              </button>
            </div>
            <div className="aspect-square rounded-xl overflow-hidden border border-border">
              <img src={previewItem.url} alt={previewItem.name} className="w-full h-full object-contain bg-black/5" />
            </div>
            <div className="text-xs text-muted-foreground space-y-1">
              <div><strong>Size:</strong> {previewItem.size}</div>
              <div><strong>Dimensions:</strong> {previewItem.dimensions}</div>
              <div><strong>Uploaded:</strong> {previewItem.uploadedAt}</div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
