import { AdminLayout } from "../../components/admin/AdminLayout";
import { VendorProductsEditor } from "../../components/admin/VendorProductsEditor";

export function VendorProductsPage() {
  return <AdminLayout title="我的产品资料">
    <VendorProductsEditor />
  </AdminLayout>;
}
