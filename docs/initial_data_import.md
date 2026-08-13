# Initial Product Data Import

The `initial_data_import` Odoo addon is installed for every new tenant. Stock managers open **Inventory > Configuration > Initial Data Import**.

## Supported files

- CSV encoded as UTF-8, with comma, semicolon, or tab delimiter detection
- XLSX, using the active worksheet
- Header row within the first ten rows

The header must contain **Machine Code** and **Description**. Recognized columns are Machine Code, Model Code, Description, PRICE ETB, PRICE USD, Picture columns, and STOCK columns. Header matching ignores case, spaces, and punctuation. The importer reads at most STOCK-1, STOCK-2, and STOCK-3 into the preview/import model.

## Import behavior

- Description is required
- Machine Code is optional but strongly recommended
- A matching Machine Code updates the existing product; otherwise a product is created
- Products are made active and storable
- ETB becomes the product sales price and an ETB pricelist fixed price when nonzero
- USD creates or updates a USD pricelist fixed price when nonzero
- ETB currency is created if missing; USD must already exist
- STOCK-1 through STOCK-3 become internal child locations under the first company warehouse
- Each stock value is treated as the target quantity, not an amount to add
- Negative stock values and invalid numbers block import

## Pictures

Only the first embedded XLSX image found in a recognized Picture column for a row is imported into `image_1920`. CSV picture values, file paths, and text URLs are ignored. Although four picture columns can be recognized, the product receives one primary image.

## Safe workflow

1. Retain an unchanged customer source file.
2. Upload and select **Preview**.
3. Review every error and warning.
4. Deselect rows that should not import.
5. Record expected product and stock totals.
6. Import in a disposable/test tenant first.
7. Compare products, pricelists, locations, and quantities.

## Known validation gap

The feature has unit-level implementation coverage but still needs acceptance testing against a representative customer workbook, especially merged/two-row headers and embedded pictures.
