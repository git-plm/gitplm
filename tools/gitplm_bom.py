#
# Example python script to generate a BOM from a KiCad generic netlist
#
# Example: Sorted and Grouped CSV BOM
#

"""
    @package
    Output: CSV (comma-separated)
    Grouped By: Value, Footprint
    Sorted By: Ref
    Fields: Ref, Qnty, Value, Cmp name, Footprint, Description, Vendor

    Command line:
    python "pathToFile/bom_csv_grouped_by_value_with_fp.py" "%I" "%O.csv"
"""

# Import the KiCad python helper module and the csv formatter
import kicad_netlist_reader
import csv
import sys

# Generate an instance of a generic netlist, and load the netlist tree from
# the command line option. If the file doesn't exist, execution will stop
net = kicad_netlist_reader.netlist(sys.argv[1])

# Open a file to write to, if the file cannot be opened output to stdout
# instead

# look for PCB-NNN.yml file in output directory and then name output file to
# PCB-NNN.csv. If PCB-NNN.csv file exists, then replace it.

from pathlib import Path

bomFile = Path(sys.argv[2])
bomDir = bomFile.parent

ymlFiles = sorted(Path(bomDir).glob('PCB-*.yml'))
if len(ymlFiles) == 1:
    print("Found yml file: ", ymlFiles[0].name)
    bomFile = bomDir / Path(ymlFiles[0].stem + ".csv")
else:
    # look for existing BOM and overwite it
    pcbCsvFiles = sorted(Path(bomDir).glob('PCB-*.csv'))
    if len(pcbCsvFiles) == 1:
        print("Found existing bom file: ", pcbCsvFiles[0].name)
        bomFile = pcbCsvFiles[0]
    else:
        print()
        print("WARNING: Did not find yml config file, please create a "
              + "PCB-NNN.yml config file so we "
              + "know the part number for this PCB.")
        print()

print("Outputing bomfile: ", str(bomFile))

try:
    f = open(str(bomFile), 'w')
except IOError:
    e = "Can't open output file for writing: " + sys.argv[2]
    print(__file__, ":", e, sys.stderr)
    f = sys.stdout

# Create a new csv writer object to use as the output formatter
out = csv.writer(f, lineterminator='\n', delimiter=';', quotechar='\"', quoting=csv.QUOTE_ALL)

# Output a set of rows for a header providing general information
# don't write header info as that breaks automatic parsing by CSV tools
#out.writerow(['Source:', net.getSource()])
#out.writerow(['Date:', net.getDate()])
#out.writerow(['Tool:', net.getTool()])
#out.writerow( ['Generator:', sys.argv[0]] )
#out.writerow(['Component Count:', len(net.components)])
out.writerow(['Ref', 'Qnty', 'Value', 'Cmp name', 'Footprint', 'Description',
              'Vendor', 'IPN', 'Datasheet'])

# Get all of the components in groups of matching parts + values
# (see ky_generic_netlist_reader.py)
grouped = net.groupComponents()

# Output all of the component information
for group in grouped:
    refs = ""

    # Add the reference of every component in the group and keep a reference
    # to the component so that the other data can be filled in once per group
    for component in group:
        refs += component.getRef() + ", "
        c = component

    # Fill in the component groups common data
    out.writerow([refs, len(group), c.getValue(), c.getPartName(), c.getFootprint(),
        c.getDescription(), c.getField("Vendor"), c.getField("IPN"), c.getDatasheet()])
