import { AdminLayout } from "../../components/admin/AdminLayout";
import { VendorProductsEditor } from "../../components/admin/VendorProductsEditor";
import { Camera } from "lucide-react";
import { Link } from "react-router-dom";

export function VendorProductsPage() {
  return <AdminLayout title="我的产品资料">
    <div className="vendor-capture-entry"><div><strong>纸质彩页快速录入</strong><span>拍摄名片或产品彩页，自动生成待确认产品草稿。</span></div><Link className="primary-btn" to="/admin/capture"><Camera size={16} />拍照识别</Link></div>
    <VendorProductsEditor />
  </AdminLayout>;
}
