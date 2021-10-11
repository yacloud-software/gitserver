#!/usr/bin/python
'''
    A python script example to create various plot files from a board:
    Fab files
    Doc files
    Gerber files

    Important note:
        this python script does not plot frame references.
        the reason is it is not yet possible from a python script because plotting
        plot frame references needs loading the corresponding page layout file
        (.wks file) or the default template.

        This info (the page layout template) is not stored in the board, and therefore
        not available.

        Do not try to change SetPlotFrameRef(False) to SetPlotFrameRef(true)
        the result is the pcbnew lib will crash if you try to plot
        the unknown frame references template.

The online "documentation" is here:
http://docs.kicad-pcb.org/doxygen-python/classplugins_1_1FootprintWizardBase_1_1FootprintWizard.html

Source code at:
http://docs.kicad-pcb.org/doxygen-python/pcbnew_8py_source.html

'''

import sys

from pcbnew import *
filename=sys.argv[1]
outdir=sys.argv[2]

board = LoadBoard(filename)

pctl = PLOT_CONTROLLER(board)

popt = pctl.GetPlotOptions()

popt.SetOutputDirectory(outdir)

# Set some important plot options:
popt.SetPlotFrameRef(False)
popt.SetLineWidth(FromMM(0.35))

popt.SetAutoScale(False)
popt.SetScale(1)
popt.SetMirror(False)
popt.SetUseGerberAttributes(True)
popt.SetExcludeEdgeLayer(False)
popt.SetScale(1)
popt.SetUseAuxOrigin(False)

# This by gerbers only (also the name is truly horrid!)
popt.SetSubtractMaskFromSilk(False) #remove solder mask from silk to be sure there is no silk on pads

# plot the drill files

ewrt = EXCELLON_WRITER(board)
ewrt.SetFormat(True)
ewrt.CreateDrillandMapFilesSet(outdir,True,True)

all = True

# we plot ALL layers....
for layer_id in range (0,PCB_LAYER_ID_COUNT):
    if (all or board.IsLayerEnabled(layer_id)):
        layername = board.GetLayerName(layer_id)
        print "LayerID: "+str(layer_id)+" = "+layername
        pctl.SetLayer(layer_id)
        pctl.OpenPlotfile(layername, PLOT_FORMAT_GERBER, layername)
        pctl.PlotLayer()
        pctl.OpenPlotfile(layername, PLOT_FORMAT_PDF, layername)
        pctl.PlotLayer()
        pctl.OpenPlotfile(layername, PLOT_FORMAT_SVG, layername)
        pctl.PlotLayer()
        
# At the end you have to close the last plot, otherwise you don't know when
# the object will be recycled!
pctl.ClosePlot()

# now generate the pick & place data from the board file
MODULE_ATTR_NORMAL = 0
MODULE_ATTR_NORMAL_INSERT = 1
MODULE_ATTR_VIRTUAL = 2

fp=open(outdir+'/pick_and_place.csv','w')

# get the auxiliary origin so that gerbers and pick-and-place agree
(off_x, off_y) = board.GetAuxOrigin()

for module in board.GetModules():
    if (module.GetAttributes() == MODULE_ATTR_NORMAL_INSERT):
        
        (pos_x, pos_y) = module.GetPosition()
        
        side = 'top'
        if module.IsFlipped():
            side = 'bottom'
        data = {'Reference': module.GetReference(),
                # correct for offset - note Y is swapped 
                'PosX': (pos_x-off_x)/1000000.0,
                'PosY': (off_y-pos_y)/1000000.0,
                'Rotation': module.GetOrientation()/10.0,
                'Side': side
                }
        fp.write('{0[Reference]},{0[PosX]}, {0[PosY]}, {0[Rotation]}, {0[Side]}\n'.format(data))
fp.close()
