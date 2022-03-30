# Part number guidelines

Basic format: CCC-NNN-VVVV

- CCC: major category (RES, CAP, DIO, etc)
- NNN: incrementing sequential number for each part
- VVVV: variation to code variations of a parts typically with the **same
  datasheet** (resistance, capacitance, regulator voltage, IC package, etc.).
  VVVV is also used to encode version for manufactured parts or assemblies.

Why use an intelligent PN scheme over a random or simple incrementing part
number, which is
[recommended](https://www.buyplm.com/plm-good-practice/part-numbering-system-software.aspx)
by most in the industry? There are several reasons:

- the NNN increments, so you do have a "Random" part, which gives you any
  flexibility you need.
- the CCC naturally sorts parts for you -- on your BOM, in the warehouse, in the
  factory, in your lab stock, etc. When physically dealing with 100's of part
  numbers on a large PCB, any organization is helpful.
- the VVVV allows you to group variations of the same part together. Again, when
  I go to my lab stock, if I need a 0603 resistor, I can quickly find it, versus
  having to look for a random part number. If you need to add additional values
  later, they are still grouped together vs being located 1000's of numbers
  later.
- an argument could be made that the description is adequate. The description
  works fine in electronic tools, but not as well for the parts bin in the lab.
  It is much easier to sort parts by CCC-NNN-VVVV than a long description, even
  if the description is somewhat consistent.
- an argument can be made that if you can't encode all parameters in the PN,
  then you should not encode any. I don't buy that. CCC/VVVV organization is
  very useful when dealing with physical bags of parts in the real world. To
  find commonly used parts, you will quickly memorize the small NNN number.

The CCC-NNN-VVVV scheme is a pragmatic compromise between random part numbers
and extensive descriptions where every last parameter is encoded in the
description. It is kind of like using colors on the factory floor -- you can't
encode everything in colors, but what you can encode sure helps with rapid
processing by humans.

CCC-NNN-VVVV also follows the general to specific naming convention, which is
generally the best way to name things.

If you are worried about running out of CCC/NNN variations, then make it
CCCC-NNNN-VVVV. Still very easy to manually parse on the floor.

CCC-NNN-VVVV works really well for electronic parts. It may not be optimal for
other parts, but there is nothing stopping you from using a different format for
other classes of parts.

Each group of CCC parts is placed in its own schematic symbol library with the
same name. The following CCC groups are suggested for electrical parts:

- ANA: op-amps, comparators, A/D, D/A
- CAP: capacitors
- CON: connectors
- DIO: diodes
- IND: inductors
- ICS: integrated circuits, mcus, etc
- OSC: oscillators
- PWR: relays, etc
- REG: regulators
- RES: resistors
- SWI: switch
- TRA: transistors, FETs
- TXT: test points/pads
- XTL: crystals

The following CCC groups are suggested for high level assemblies:

- PCB: Printed Circuit board
- ASY: Assembly (can be top level or subassembly -- typically represented by BOM
  and documentation)

Some additional guidelines:

- Every part number has the same number of characters in it (3-3-4). This makes
  sorting simpler and less chance of error.
- Character set is restricted to capital letters, digits, and hyphen.
- Avoid punctuation characters such as %, !, (, ., etc.

Examples:

With resistors, capacitors, and connectors, we encode the value and pin count in
the variation:

- 1K 0805 1%: RES-002-1001
- 3.3K 0805 1%: RES-002-3301
- 2.2K 0603 5%: RES-003-2201 (note we bumped NNN to 003, because different
  package size)
- 10.3K high power 0603: RES-004-1032 (different vendor/datasheet than RES-002,
  so we bump NNN)
- 2x10, 0.1 in header: CON-000-0020
- 2x12, 0.1 in header: CON-000-0024
- 1x10, 0.1 in header: CON-001-0010
- 1x20, 0.1 in header: CON-001-0020

With most ICs we simply enumerate all the variations of a particular IC in a
sequentially incrementing variation (we don't try to encode information)

- LM78xx SOT223 5V: REG-089-0000
- LM78xx DIP 5V: REG-089-0001
- LM78xx SOT223 3.3V: REG-089-0002
- LM78xx DIP 3.3V: REG-089-0003
- 3.3v switching reg, SSOP8: REG-002-0000
- 3.3v switching reg, S08: REG-002-0001
- STM32H7 in 44 pin package, 1M flash: MCU-001-0000
- STM32H7 in 44 pin package, 2M flash: MCU-001-0001
- STM32H7 in 208 pin package, 1M flash: MCU-001-0002
- STM32H7 in 208 pin package, 2M flash: MCU-001-0003
- STM32F3 in 44 pin package: MCU-002-0044 (not different base part, so bump NNN)

Many parts will not have any variations:

- 2N4401 DIODE: DIO-000-0000 (no variation information, that is fine)
- 2N2222 transistor: TRA-000-0000 (again, no variation info)

The variation section is only used in cases where a part with a single datasheet
has multiple variations. Variations are generally used to encode one parameter
with the most different variations -- for instance resistance with resistors. A
single datasheet may include 0603, 0805, and 1206 but for resistors, but take
out separate NNN part numbers for different package sizes because with
resistors, it makes the most sense to encode the resistance in the variation
(because there are lots of resistance values), not the package size. There are
relatively few package sizes for resistors so it makes sense to take out new NNN
numbers for different packages. However, for voltage regulators, it may make
sense to encode both the regulated voltage and the package in the variation,
because there is a relatively small number of combinations.

Generally we don't need to create house part numbers for every part variation --
only the ones we use. Resistors/caps may be an exception where we simply create
the entire series in the partmaster because it is easiest to just do once.

### Resistor part numbers

Most resistor variations (at least 1%) are encoded using the E96 4-digit
industry standard. Examples:

- 2500 = 250 x 100 = 250 x 1 = 250 Ω (This is only and only 250Ω not 2500 Ω)
- 1000 = 100 x 100 = 100x 1 = 100 Ω
- 7201 = 720 x 101 = 720 x 10 = 7200 Ω or 7.2kΩ
- 1001 = 100 × 101 =100 x 10 = 1000 Ω or 1kΩ
- 1004 = 100 × 104 =100 x 10000 = 1,000,000 Ω or 1MΩ
- R102 = 0.102 Ω (4-digit SMD resistors (E96 series)
- 0R10 = 0.1 x 100 = 0.1 x 1 = 0.1 Ω (4-digit SMD resistors (E24 series)
- 25R5 = 25.5Ω (4-digit SMD resistors (E96 series))

### Capacitor part numbers

Most capacitors values are encoded in a 3-digit number where the 1st two digits
are the value and the last digit is the number of zeros in pF. Since we have 4
digits, the 1st digit is typically not used, but can if precision caps exist
that need 3 significant places to encode the value. The goal is to match what
most vendors are doing so we can easily compare IPN and vendor part numbers.

Examples:

- 103 = 10 \* 10^3 = 10,000pF = 10nF = 0.01uF
- 104 = 0.1uF

To figure out the extension, you can divide the capitance by 1pF to get the
number of pF. From this, you can visually tell what the variation should be.
Example:

`0.022uF = 0.022e-6/1e-12 = 22000`

So the extension would be `223`

To work backwards, we would have `1000 * 22 = 22,000pF/1e-6 = 0.022uF`.
