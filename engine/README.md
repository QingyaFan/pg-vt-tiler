# engine

Engine is for generating vector tiles in every specified zoom levels within data extent. To do that, we need three steps:

1. get x y boundry from data extent and zoom level
2. loop over every tile zxy in the boundry, calculate it's extent and get the data in this extent
3. generate the tile, and write to file
