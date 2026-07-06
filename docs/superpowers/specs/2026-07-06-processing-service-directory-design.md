# Processing Service Directory Design

## Goal

Add a dedicated processing service directory for vendors that can take machining, customization, or contract processing orders. The directory should reuse the existing vendor profile as the source of truth, while adding processing capability fields and processing-specific tags.

The public `/service` page will show which vendors provide processing services, what capabilities they offer, and how buyers can enter the vendor profile to create an inquiry opportunity.

## Scope

- Reuse the existing `vendors` records for company identity, region, contact, verification, website URL, and vendor detail pages.
- Add processing service fields to vendors instead of creating an independent processing service table.
- Add processing service tags through the existing `tags` table by using a new tag type.
- Replace the current `/service` content-only page with a real processing vendor directory.
- Extend the admin vendor form so platform operators can register and maintain processing capabilities for the same vendor record.

Out of scope for this iteration:

- Order management, quotation workflow, payment, messaging, and service-level agreements.
- Separate self-service registration forms for vendors.
- Multiple independent service listings under one vendor.

## Data Model

Extend `Vendor` with processing service fields:

- `providesProcessing`: whether the vendor should appear in the processing service directory.
- `processingServices`: summary of offered processing capabilities.
- `processingMaterials`: supported materials or part types.
- `processingEquipment`: key equipment, production lines, or process assets.
- `processingCapacity`: capacity, batch size, lead time, or delivery notes.
- `processingRegions`: service coverage or logistics regions.
- `processingNotes`: cooperation notes, drawing/sample requirements, or order preferences.

Extend tag usage by allowing `Tag.TagType = "processing"`. Processing tags are separate from normal vendor tags, but still attach to vendors through the existing `vendor_tags` relationship. Examples:

- Drawing/sample processing
- CNC turning
- Welding
- Sheet metal cutting
- Heat treatment
- Surface treatment
- Small-batch trial production
- Batch OEM processing

Existing vendor tags keep using `TagType = "vendor"`.

## Public API

Add processing-specific public endpoints:

- `GET /api/processing-vendors`
  - Returns paginated vendors where `is_visible = true` and `provides_processing = true`.
  - Supports `keyword`, `province`, `tagId`, `sort`, `page`, and `pageSize`.
  - `keyword` searches vendor name, short name, main products, processing services, materials, equipment, and notes.
  - `tagId` filters by processing tags.
  - Default sort follows recommended vendors first, then `sort_order`, then `id`.

- `GET /api/processing-filter-options`
  - Returns provinces from processing-enabled visible vendors.
  - Returns tags where `tag_type = "processing"`.

The existing `/api/vendors` endpoint remains focused on the general vendor directory.

## Admin API

Reuse existing admin endpoints:

- `GET /api/admin/vendors`
- `POST /api/admin/vendors`
- `PUT /api/admin/vendors/:id`
- `GET /api/admin/tags`
- `POST /api/admin/tags`
- `PUT /api/admin/tags/:id`

The admin vendor save flow should persist the new processing fields. Tag validation should continue to verify that selected tag IDs exist. The form can allow both vendor and processing tags to attach to the same vendor, while public pages decide which tag type to display.

## Frontend

Create a dedicated `ProcessingServicesPage` for `/service`.

The page should follow the current `VendorsPage` pattern:

- Page title: `加工服务`
- Subtitle: explain that vendors can register processing capabilities and receive matching processing order opportunities.
- Filters:
  - keyword search for vendor name, processing service, material, equipment, or region
  - province selector
  - processing service tag selector
  - sort selector
- Result summary: total number of processing vendors found.
- Grid/list of vendor cards.
- Empty, loading, and error states consistent with the vendor directory.

Cards should reuse the existing visual language, but highlight processing fields:

- vendor name and region
- processing tags
- processing services
- processing equipment or capacity
- primary action links to `/vendors/:id`

The normal vendor detail page should include a processing service section when `providesProcessing` is true. This keeps the final buyer path on the same vendor profile and avoids duplicate detail pages.

## Admin Frontend

Extend the existing "厂商信息" form with processing service fields:

- 是否提供加工服务
- 加工服务能力
- 可加工材料 / 配件类型
- 加工设备
- 产能 / 交期
- 服务区域
- 接单说明
- 厂商标签, including processing tags

Extend tag type options from vendor/product to vendor/product/processing.

## Seed Data

Add representative processing tags and mark several seeded vendors as processing-capable. Seeded vendors should include varied service capabilities so `/service` works immediately after database initialization.

## Testing

Backend:

- Processing vendor API filters only visible vendors with `providesProcessing = true`.
- Keyword search matches processing fields.
- Processing tag filter returns matching vendors.
- Processing filter options return only processing tags.
- Admin vendor create/update persists processing fields and tag IDs.

Frontend:

- `/service` renders title, filters, result summary, and processing vendor cards.
- Filters call the processing API with expected params.
- Empty and error states render correctly.
- Admin vendor form exposes processing fields.
- Tag form supports the processing tag type.

## Acceptance Criteria

- `/service` is a real directory page, not a static content page.
- A vendor can be maintained once in the admin and appear in both the general vendor directory and the processing service directory when enabled.
- Processing service vendors can be filtered by keyword, region, and processing capability tag.
- Processing service cards show useful service capability information, not only generic company information.
- Vendor detail pages show processing service information for processing-capable vendors.
