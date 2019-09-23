# Generator

The generator is for generating vector tiles in every specified zoom levels within data extent. To do that, we need three steps:

1. compute data extent, and then get the zxy range
1. for every zxy, translate zxy to zoom and extent
1. generate the tile, and write to file
